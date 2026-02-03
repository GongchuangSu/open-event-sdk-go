// Package model 定义事件数据模型
// 该包中的类型定义可通过代码生成器自动生成
// 类型命名与 xidl 定义保持一致，使用 V7 前缀
package model

// V7Identity 身份唯一标识
type V7Identity struct {
	// Id 身份ID
	Id string `json:"id"`

	// Type 身份类型
	// 可选值: user(用户), sp(服务主体), app(应用), unknown(未知)
	Type string `json:"type"`

	// Name 用户或应用的名称
	Name string `json:"name,omitempty"`

	// Avatar 用户或应用的头像
	Avatar string `json:"avatar,omitempty"`

	// CompanyId 身份所归属的公司
	CompanyId string `json:"company_id,omitempty"`
}

// V7IdentityType 身份类型常量
const (
	V7IdentityTypeUser    = "user"    // 用户
	V7IdentityTypeSp      = "sp"      // 服务主体
	V7IdentityTypeApp     = "app"     // 应用
	V7IdentityTypeUnknown = "unknown" // 未知
)

// V7ExtendedAttribute 扩展属性
type V7ExtendedAttribute struct {
	// Name 属性名
	Name string `json:"name"`

	// Value 属性值
	Value string `json:"value"`
}
