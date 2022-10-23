package form

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/yaoapp/gou"
)

// BindModel bind model
func (fields *FieldsDSL) BindModel(m *gou.Model) {

	// trans, err := field.ModelTransform()
	// if err != nil {
	// 	return
	// }

	// for _, col := range m.Columns {
	// 	data := col.Map()
	// 	tableField, err := trans.Table(col.Type, data)
	// 	if err != nil {
	// 		return
	// 	}
	// 	// append columns
	// 	if _, has := fields.Form[tableField.Key]; !has {
	// 		fields.Form[tableField.Key] = *tableField
	// 		// fields.tableMap[col.Name] = fields.Table[tableField.Key]
	// 	}
	// }
}

// Xgen trans to xgen setting
func (fields *FieldsDSL) Xgen() (map[string]interface{}, error) {
	res := map[string]interface{}{}
	data, err := jsoniter.Marshal(fields)
	if err != nil {
		return nil, err
	}

	err = jsoniter.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
