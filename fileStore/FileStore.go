package fileStore

import (
	"crypto"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	myHelper "github.com/livegoplayer/go_helper"

	"github.com/livegoplayer/filestore-server/model"
)

//默认值
var DEFAULT_PATH string

func init() {
	DEFAULT_PATH = ""
}

//这样定义比较好
var (
	err error
)

//根据二进制文件保存文件
func SaveFileToDir(file multipart.File, newFileName string, toPath string, fileSha1 string) (*FileMeta, error) {
	//解析文件后缀，分别放到不同的文件夹
	var fileMeta = &FileMeta{}
	//如果选择使用默认路径，默认都是默认路径
	if toPath == DEFAULT_PATH {
		toPath = GetDefaultPath(newFileName)
	}

	//创建新文件
	createdFile, err := os.Create(path.Join(toPath, "/", newFileName))
	if err != nil {
		return fileMeta, err
	}
	defer createdFile.Close()

	//复制文件内容到新文件
	fileSize, err := io.Copy(createdFile, file)
	//复制文件
	if err != nil {
		return fileMeta, err
	}

	//初始化文件元信息
	fileMeta.FileName = newFileName
	fileMeta.FileSize = fileSize
	//文件指针重置
	_, _ = createdFile.Seek(0, 0)
	fileMeta.FileSha1 = myHelper.FileSha1(createdFile)
	fileMeta.Location = myHelper.PathToCommon(toPath + "/" + newFileName)
	fileMeta.Type = GetFileTypeByName(newFileName)
	fileMeta.UploadTime = time.Now()
	fileMeta.UpdateTime = time.Now()
	fileMeta.FileSha1 = fileSha1

	return fileMeta, nil
}

//为用户增加一个文件
func AddFileToUser(fileHeader *multipart.FileHeader, newFileName string, toPath string, uid int, pathId int) (fileMeta *FileMeta, err error) {
	fileMeta = nil
	err = nil

	file, err := myHelper.GetFileByHeader(fileHeader)
	if err != nil {
		return
	}
	fileSha1 := myHelper.FileSha1(file)

	//如果该文件存在
	var fileId = 0
	if fileModel, exist := model.CheckFileExist(fileSha1); exist {
		fileMeta = GetFileMetaByFile(fileModel)
		//根据fileMeta获取file信息
		fileId = fileModel.Id
	} else {
		//重新获取file对象，因为file被sha1方法破坏了
		file, err = myHelper.GetFileByHeader(fileHeader)
		if err != nil {
			return
		}
		fileMeta, err = SaveFileToDir(file, newFileName, toPath, fileSha1)
		if err != nil {
			return
		}
		fileId = model.SaveFileToMysql(fileMeta.FileSha1, fileMeta.Location, fileMeta.FileSize, model.LocalStore, "")
	}

	if fileId > 0 {
		_ = model.SaveFileToUser(fileId, newFileName, uid, pathId, fileMeta.FileSize, fileMeta.Type)
	}

	return fileMeta, err
}

/**
bucketName OSS上存储空间的名字
fileOSSName OSS服务器上的文件名
fileOSSPath OSS服务器上的路径
fileSize 文件大小
fileSha1 文件唯一校验码
fileSha1 文件唯一校验码
*/
func AddOSSFileToUser(bucketName, fileOSSName, filename, fileOSSPath, fileSha1 string, uid, pathId int, fileSize int64) int {
	//第一步，保存到file表中
	fileId := model.SaveFileToMysql(fileSha1, fileOSSPath+"/"+fileOSSName, fileSize, model.OSSStore, bucketName)

	id := model.SaveFileToUser(fileId, filename, uid, pathId, fileSize, GetFileTypeByName(fileOSSName))

	return id
}

func AddExistOSSFileToUser(fileId int, fileOSSName string, uid, pathId int, fileSize int64) int {
	id := model.SaveFileToUser(fileId, fileOSSName, uid, pathId, fileSize, GetFileTypeByName(fileOSSName))

	return id
}

//单独操作文件夹
func SaveUserFilePath(userPath *model.UserPath) model.UserPath {
	if userPath.ID > 0 {
		return model.UpdateUserPath(*userPath, []int{})
	}
	return model.AddUserPath(*userPath)
}

//批量操作文件夹
func UpdateUserPath(ipMap []int, userPath model.UserPath) bool {
	model.UpdateUserPath(userPath, ipMap)
	return true
}

//批量操作文件夹
func UpdateUserFile(ipMap []int, retUserFile model.RetUserFile) bool {
	model.UpdateUserFile(retUserFile, ipMap)
	return true
}

func DelUserPath(idMap []int) bool {

	//todo 事务操作
	model.UpdateUserPath(model.UserPath{Status: 9}, idMap)

	//删除文件夹中的文件
	success := model.DelFilesInPath(idMap)

	return success
}

func DelUserFile(idMap []int) bool {
	model.UpdateUserFile(model.RetUserFile{Status: 9}, idMap)
	return true
}

//根据文件后缀名获取默认存储路径
func GetDefaultPath(fileName string) string {
	ext := myHelper.GetFileExtName(fileName)
	var Path string
	if ext == "" {
		Path = path.Join("./files/", "unknown", "/")
	} else {
		Path = path.Join("./files/", ext, "/")
	}

	defaultSavePath := myHelper.PathToCommon(Path)

	//确保文件夹已经存在
	err := os.MkdirAll(defaultSavePath, 0666)
	//如果创建出错
	if err != nil {
		panic(err)
	}

	return defaultSavePath
}

func CheckFileExists(fileSha1 string) (*model.File, bool) {
	return model.CheckFileExist(fileSha1)
}

func GetFileListByPathId(uid int, pathId int, searchKey string) []model.RetUserFile {
	return model.GetFileListByPath(uid, pathId, searchKey)
}

func GetUserPathList(uid int) []model.UserPath {
	return model.GetUserPathList(uid)
}

func GetUserChildPathList(uid int, pid int, searchKey string) []model.UserPath {
	return model.GetUserChildPathList(uid, pid, searchKey)
}

//递归函数
func GetChildPathIdList(pid int, list []model.UserPath, parentIdList []int) (idList []int) {
	idList = []int{pid}

	//初始化parentIdList，这个递归函数中该函数只执行一次
	if len(parentIdList) == 0 {
		//先拿到所有的parent_id-id map
		for _, oneUserPath := range list {
			parentIdList = append(parentIdList, oneUserPath.ParentId)
		}
	}

	//递归
	for _, oneUserPath := range list {
		if oneUserPath.ParentId == pid {
			idList = append(idList, oneUserPath.ID)
			//如果有节点以此节点作为parent_id
			if exists, _ := myHelper.InArray(oneUserPath.ID, parentIdList); exists {
				childList := GetChildPathIdList(oneUserPath.ID, list, parentIdList)
				idList = append(idList, childList...)
			}
		}
	}
	return idList

}

//根据file获取初始化好的FileMeta对象 todo 增加user file对象
func GetFileMetaByFile(file *model.File) *FileMeta {
	fileMeta := &FileMeta{}
	fileMeta.FileSha1 = file.FileSha1
	fileMeta.FileSize = file.Size
	fileMeta.Location = file.Path
	fileMeta.Type = GetFileTypeByName(file.Path)
	_, fileMeta.FileName = path.Split(file.Path)
	fileMeta.UpdateTime = myHelper.Str2Time(file.UpdateDatetime)
	fileMeta.UploadTime = myHelper.Str2Time(file.AddDatetime)

	return fileMeta
}

const (
	FOLDER = 1
	IMG    = 2
	VIDEO  = 3
	OTHER  = 4
	PDF    = 5
)

func GetFileTypeByName(filename string) int {
	if match, _ := regexp.MatchString("(.*)\\.(jpg|bmp|gif|ico|pcx|jpeg|tif|png|raw|tga)", filename); match {
		return IMG
	}

	if match, _ := regexp.MatchString("(.*)\\.(swf|flv|mp4|rmvb|avi|mpeg|ra|ram|mov|wmv)", filename); match {
		return VIDEO
	}

	if match, _ := regexp.MatchString("(.*)\\.(pdf)", filename); match {
		return PDF
	}

	return OTHER
}

/*************************************************************************************************************************************/

// service/upload_token_service.go
type UpLoadToOSSService struct {
	FileName string `form:"filename" json:"filename" binding:"required"`
}

var ossClient *oss.Client

/**
<yourObjectName>上传文件到OSS时需要指定包含文件后缀在内的完整路径，例如abc/efg/123.jpg。
<yourLocalFileName>由本地文件路径加文件名包括后缀组成，例如/users/local/myfile.txt。
*/
func UploadFileToOss(bucketName, objectName, localFileName string) {
	// 获取存储空间。
	bucket, err := ossClient.Bucket(bucketName)
	if err != nil {
		panic(err)
	}
	// 上传文件。
	err = bucket.PutObjectFromFile(objectName, localFileName)
	if err != nil {
		panic(err)
	}
}

//临时授权下载
func GetDownloadUrl(bucketName, objectName string) string {
	// 获取存储空间。
	bucket, err := ossClient.Bucket(bucketName)
	if err != nil {
		panic(err)
	}
	// 使用签名URL将OSS文件下载到流。
	signedURL, err := bucket.SignURL(objectName, oss.HTTPGet, 60)
	if err != nil {
		panic(err)
	}

	return signedURL
}

func PrepareForUpLoad(bucketName string, fileName string, pathToSave string) string {

	// 获取存储空间
	bucket, err := ossClient.Bucket(bucketName)
	if err != nil {
		panic(err)
	}

	// 获取扩展名
	ext := filepath.Ext(fileName)

	// 带可选参数的签名直传
	options := []oss.Option{
		oss.ContentType(mime.TypeByExtension(ext)),
	}

	key := pathToSave
	// 生成签名url, 签名直传，10分钟内有效
	signedPutURL, err := bucket.SignURL(key, oss.HTTPPut, 600, options...)
	if err != nil {
		panic(err)
	}

	return signedPutURL
}

//获取临时授权token
func GetPolicyToken(expireTime int64, uploadDir string, callbackParam CallbackParam, bucketName string) PolicyToken {
	now := time.Now().Unix()
	expireEnd := now + expireTime
	var tokenExpire = get_gmt_iso8601(expireEnd)

	//create post policy json
	var config ConfigStruct
	config.Expiration = tokenExpire
	var condition []string
	condition = append(condition, "starts-with")
	condition = append(condition, "$key")
	condition = append(condition, uploadDir)
	config.Conditions = append(config.Conditions, condition)

	//calucate signature
	result, err := json.Marshal(config)
	debyte := base64.StdEncoding.EncodeToString(result)
	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(ossConfig.AccessKeySecret))
	io.WriteString(h, debyte)
	signedStr := base64.StdEncoding.EncodeToString(h.Sum(nil))

	if callbackParam.CallbackUrl == "" {
		callbackParam.CallbackBody = ""
		callbackParam.CallbackBodyType = ""
	} else {
		if callbackParam.CallbackBodyType == "" {
			callbackParam.CallbackBodyType = "application/x-www-form-urlencoded"
		}
	}
	callback_str, err := json.Marshal(callbackParam)
	if err != nil {
		panic(err)
	}
	callbackBase64 := base64.StdEncoding.EncodeToString(callback_str)

	var policyToken PolicyToken
	policyToken.AccessKeyId = ossConfig.AccessKeyId
	policyToken.Host = "http://" + bucketName + "." + ossConfig.Endpoint
	policyToken.Expire = expireEnd
	policyToken.Signature = string(signedStr)
	policyToken.Directory = uploadDir
	policyToken.Policy = string(debyte)
	policyToken.Callback = string(callbackBase64)

	return policyToken
}

type ConfigStruct struct {
	Expiration string     `json:"expiration"`
	Conditions [][]string `json:"conditions"`
}

type CallbackParam struct {
	CallbackUrl      string `json:"callbackUrl"`
	CallbackBody     string `json:"callbackBody"`
	CallbackBodyType string `json:"callbackBodyType"`
}

func get_gmt_iso8601(expire_end int64) string {
	var tokenExpire = time.Unix(expire_end, 0).Format("2006-01-02T15:04:05Z")
	return tokenExpire
}

type OSSConfig struct {
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
	CallbackUrl     string
	UploadDir       string
	ExpireTime      string
}

type PolicyToken struct {
	AccessKeyId string `json:"access_id"`
	Host        string `json:"host"`
	Expire      int64  `json:"expire"`
	Signature   string `json:"signature"`
	Policy      string `json:"policy"`
	Directory   string `json:"dir"`
	Callback    string `json:"callback"`
}

var ossConfig = &OSSConfig{}

// 请填写您的AccessKeyId。
// 请填写您的AccessKeySecret。
// host的格式为 bucketname.endpoint ，请替换为您的真实信息。
// callbackUrl为 上传回调服务器的URL，请将下面的IP和Port配置为您自己的真实信息。
// 用户上传文件时指定的前缀。
func InitOSSClient(accessKeyId, accessKeySecret, endpoint string) {
	ossConfig.AccessKeyId = accessKeyId
	ossConfig.AccessKeySecret = accessKeySecret
	ossConfig.Endpoint = endpoint

	var err error
	ossClient, err = oss.New(endpoint, accessKeyId, accessKeySecret, oss.Timeout(10, 120))
	if err != nil {
		panic(err)
	}
}

// getPublicKey : Get PublicKey bytes from Request.URL
func GetPublicKey(r *http.Request) ([]byte, error) {
	var bytePublicKey []byte
	// get PublicKey URL
	publicKeyURLBase64 := r.Header.Get("x-oss-pub-key-url")
	if publicKeyURLBase64 == "" {
		fmt.Println("GetPublicKey from Request header failed :  No x-oss-pub-key-url field. ")
		return bytePublicKey, errors.New("no x-oss-pub-key-url field in Request header ")
	}

	publicKeyURL, _ := base64.StdEncoding.DecodeString(publicKeyURLBase64)
	// fmt.Printf("publicKeyURL={%s}\n", publicKeyURL)
	// get PublicKey Content from URL
	responsePublicKeyURL, err := http.Get(string(publicKeyURL))
	if err != nil {
		fmt.Printf("Get PublicKey Content from URL failed : %s \n", err.Error())
		return bytePublicKey, err
	}
	bytePublicKey, err = ioutil.ReadAll(responsePublicKeyURL.Body)
	if err != nil {
		fmt.Printf("Read PublicKey Content from URL failed : %s \n", err.Error())
		return bytePublicKey, err
	}
	defer responsePublicKeyURL.Body.Close()
	// fmt.Printf("publicKey={%s}\n", bytePublicKey)
	return bytePublicKey, nil
}

// getAuthorization : decode from Base64String
func GetAuthorization(r *http.Request) ([]byte, error) {
	var byteAuthorization []byte
	// Get Authorization bytes : decode from Base64String
	strAuthorizationBase64 := r.Header.Get("authorization")
	if strAuthorizationBase64 == "" {
		fmt.Println("Failed to get authorization field from request header. ")
		return byteAuthorization, errors.New("no authorization field in Request header")
	}
	byteAuthorization, _ = base64.StdEncoding.DecodeString(strAuthorizationBase64)
	return byteAuthorization, nil
}

// getMD5FromNewAuthString : Get MD5 bytes from Newly Constructed Authrization String.
func GetMD5FromNewAuthString(r *http.Request) ([]byte, error) {
	var byteMD5 []byte
	// Construct the New Auth String from URI+Query+Body
	bodyContent, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		fmt.Printf("Read Request Body failed : %s \n", err.Error())
		return byteMD5, err
	}
	strCallbackBody := string(bodyContent)
	// fmt.Printf("r.URL.RawPath={%s}, r.URL.Query()={%s}, strCallbackBody={%s}\n", r.URL.RawPath, r.URL.Query(), strCallbackBody)
	strURLPathDecode, errUnescape := UnescapePath(r.URL.Path, encodePathSegment) //url.PathUnescape(r.URL.Path) for Golang v1.8.2+
	if errUnescape != nil {
		fmt.Printf("url.PathUnescape failed : URL.Path=%s, error=%s \n", r.URL.Path, err.Error())
		return byteMD5, errUnescape
	}

	// Generate New Auth String prepare for MD5
	strAuth := ""
	if r.URL.RawQuery == "" {
		strAuth = fmt.Sprintf("%s\n%s", strURLPathDecode, strCallbackBody)
	} else {
		strAuth = fmt.Sprintf("%s?%s\n%s", strURLPathDecode, r.URL.RawQuery, strCallbackBody)
	}
	// fmt.Printf("NewlyConstructedAuthString={%s}\n", strAuth)

	// Generate MD5 from the New Auth String
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(strAuth))
	byteMD5 = md5Ctx.Sum(nil)

	return byteMD5, nil
}

/*  VerifySignature
*   VerifySignature需要三个重要的数据信息来进行签名验证： 1>获取公钥PublicKey;  2>生成新的MD5鉴权串;  3>解码Request携带的鉴权串;
*   1>获取公钥PublicKey : 从RequestHeader的"x-oss-pub-key-url"字段中获取 URL, 读取URL链接的包含的公钥内容， 进行解码解析， 将其作为rsa.VerifyPKCS1v15的入参。
*   2>生成新的MD5鉴权串 : 把Request中的url中的path部分进行urldecode， 加上url的query部分， 再加上body， 组合之后进行MD5编码， 得到MD5鉴权字节串。
*   3>解码Request携带的鉴权串 ： 获取RequestHeader的"authorization"字段， 对其进行Base64解码，作为签名验证的鉴权对比串。
*   rsa.VerifyPKCS1v15进行签名验证，返回验证结果。
* */
func VerifySignature(bytePublicKey []byte, byteMd5 []byte, authorization []byte) bool {
	pubBlock, _ := pem.Decode(bytePublicKey)
	if pubBlock == nil {
		panic("Failed to parse PEM block containing the public key")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if (pubInterface == nil) || (err != nil) {
		panic("x509.ParsePKIXPublicKey(publicKey) failed : " + err.Error() + "\n")
	}
	pub := pubInterface.(*rsa.PublicKey)

	errorVerifyPKCS1v15 := rsa.VerifyPKCS1v15(pub, crypto.MD5, byteMd5, authorization)
	if errorVerifyPKCS1v15 != nil {
		panic("\nSignature Verification is Failed : " + errorVerifyPKCS1v15.Error() + "\n")
	}

	fmt.Printf("\nSignature Verification is Successful. \n")
	return true
}

type encoding int

const (
	encodePath encoding = 1 + iota
	encodePathSegment
	encodeHost
	encodeZone
	encodeUserPassword
	encodeQueryComponent
	encodeFragment
)

// unescapePath : unescapes a string; the mode specifies, which section of the URL string is being unescaped.
func UnescapePath(s string, mode encoding) (string, error) {
	// Count %, check that they're well-formed.
	mode = encodePathSegment
	n := 0
	hasPlus := false
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			n++
			if i+2 >= len(s) || !Ishex(s[i+1]) || !Ishex(s[i+2]) {
				s = s[i:]
				if len(s) > 3 {
					s = s[:3]
				}
				return "", EscapeError(s)
			}
			// Per https://tools.ietf.org/html/rfc3986#page-21
			// in the host component %-encoding can only be used
			// for non-ASCII bytes.
			// But https://tools.ietf.org/html/rfc6874#section-2
			// introduces %25 being allowed to escape a percent sign
			// in IPv6 scoped-address literals. Yay.
			if mode == encodeHost && Unhex(s[i+1]) < 8 && s[i:i+3] != "%25" {
				return "", EscapeError(s[i : i+3])
			}
			if mode == encodeZone {
				// RFC 6874 says basically "anything goes" for zone identifiers
				// and that even non-ASCII can be redundantly escaped,
				// but it seems prudent to restrict %-escaped bytes here to those
				// that are valid host name bytes in their unescaped form.
				// That is, you can use escaping in the zone identifier but not
				// to introduce bytes you couldn't just write directly.
				// But Windows puts spaces here! Yay.
				v := Unhex(s[i+1])<<4 | Unhex(s[i+2])
				if s[i:i+3] != "%25" && v != ' ' && ShouldEscape(v, encodeHost) {
					return "", EscapeError(s[i : i+3])
				}
			}
			i += 3
		case '+':
			hasPlus = mode == encodeQueryComponent
			i++
		default:
			if (mode == encodeHost || mode == encodeZone) && s[i] < 0x80 && ShouldEscape(s[i], mode) {
				return "", InvalidHostError(s[i : i+1])
			}
			i++
		}
	}

	if n == 0 && !hasPlus {
		return s, nil
	}

	t := make([]byte, len(s)-2*n)
	j := 0
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			t[j] = Unhex(s[i+1])<<4 | Unhex(s[i+2])
			j++
			i += 3
		case '+':
			if mode == encodeQueryComponent {
				t[j] = ' '
			} else {
				t[j] = '+'
			}
			j++
			i++
		default:
			t[j] = s[i]
			j++
			i++
		}
	}
	return string(t), nil
}

// Please be informed that for now shouldEscape does not check all
// reserved characters correctly. See golang.org/issue/5684.
func ShouldEscape(c byte, mode encoding) bool {
	// §2.3 Unreserved characters (alphanum)
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return false
	}

	if mode == encodeHost || mode == encodeZone {
		// §3.2.2 Host allows
		//	sub-delims = "!" / "$" / "&" / "'" / "(" / ")" / "*" / "+" / "," / ";" / "="
		// as part of reg-name.
		// We add : because we include :port as part of host.
		// We add [ ] because we include [ipv6]:port as part of host.
		// We add < > because they're the only characters left that
		// we could possibly allow, and Parse will reject them if we
		// escape them (because hosts can't use %-encoding for
		// ASCII bytes).
		switch c {
		case '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', ':', '[', ']', '<', '>', '"':
			return false
		}
	}

	switch c {
	case '-', '_', '.', '~': // §2.3 Unreserved characters (mark)
		return false

	case '$', '&', '+', ',', '/', ':', ';', '=', '?', '@': // §2.2 Reserved characters (reserved)
		// Different sections of the URL allow a few of
		// the reserved characters to appear unescaped.
		switch mode {
		case encodePath: // §3.3
			// The RFC allows : @ & = + $ but saves / ; , for assigning
			// meaning to individual path segments. This package
			// only manipulates the path as a whole, so we allow those
			// last three as well. That leaves only ? to escape.
			return c == '?'

		case encodePathSegment: // §3.3
			// The RFC allows : @ & = + $ but saves / ; , for assigning
			// meaning to individual path segments.
			return c == '/' || c == ';' || c == ',' || c == '?'

		case encodeUserPassword: // §3.2.1
			// The RFC allows ';', ':', '&', '=', '+', '$', and ',' in
			// userinfo, so we must escape only '@', '/', and '?'.
			// The parsing of userinfo treats ':' as special so we must escape
			// that too.
			return c == '@' || c == '/' || c == '?' || c == ':'

		case encodeQueryComponent: // §3.4
			// The RFC reserves (so we must escape) everything.
			return true

		case encodeFragment: // §4.1
			// The RFC text is silent but the grammar allows
			// everything, so escape nothing.
			return false
		}
	}

	// Everything else must be escaped.
	return true
}

func Ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

func Unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

type EscapeError string

func (e EscapeError) Error() string {
	return "invalid URL escape " + strconv.Quote(string(e))
}

type InvalidHostError string

func (e InvalidHostError) Error() string {
	return "invalid character " + strconv.Quote(string(e)) + " in host name"
}
