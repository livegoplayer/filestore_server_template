package model

type User struct {
	Model
	ID             int    `gorm:"column:uid"` //会被自动认为是主键
	Name           string `gorm:"column:username"`
	Password       string
	AddDatetime    int `gorm:"column:add_datetime"`
	UpdateDatetime int `gorm:"column:update_datetime"`
}

//设置表名，可以通过给struct类型定义 TableName函数，返回当前struct绑定的mysql表名是什么
func (u *User) TableName() string {
	//绑定MYSQL表名为users
	return prefix + "users"
}

type Role struct {
	Model
	ID             int    //会被自动认为是主键
	Name           string `gorm:"column:role_name"`
	Level          int    `gorm:"column:role_level"`
	AddDatetime    int    `gorm:"column:add_datetime"`
	UpdateDatetime int    `gorm:"column:update_datetime"`
}

//设置表名，可以通过给struct类型定义 TableName函数，返回当前struct绑定的mysql表名是什么
func (u *Role) TableName() string {
	//绑定MYSQL表名为users
	return prefix + "role"
}

type Authority struct {
	Model
	ID   int    //会被自动认为是主键
	Name string `gorm:"column:authority_name"`
}

//设置表名，可以通过给struct类型定义 TableName函数，返回当前struct绑定的mysql表名是什么
func (u *Authority) TableName() string {
	//绑定MYSQL表名为users
	return prefix + "role"
}

type RetRoleAuthority struct {
	ID          int //会被自动认为是主键
	RoleId      int `gorm:"column:role_id"`
	AuthorityId int `gorm:"column:authority_id"`
}

//设置表名，可以通过给struct类型定义 TableName函数，返回当前struct绑定的mysql表名是什么
func (u *RetRoleAuthority) TableName() string {
	//绑定MYSQL表名为users
	return prefix + "ret_role_authority"
}

type RetUserRole struct {
	ID     int //会被自动认为是主键
	UserId int `gorm:"column:uid"`
	RoleId int `gorm:"column:role_id"`
}

//设置表名，可以通过给struct类型定义 TableName函数，返回当前struct绑定的mysql表名是什么
func (u *RetUserRole) TableName() string {
	//绑定MYSQL表名为users
	return prefix + "ret_user_role"
}

func init() {
	// 需要在init中注册定义的model fs_user
}
