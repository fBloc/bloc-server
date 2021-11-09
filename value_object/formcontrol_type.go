package value_object

// FormControlType 前端控件的类型
type FormControlType string

const (
	InputFormControl    FormControlType = "input"
	SelectFormControl   FormControlType = "select"
	RadioFormControl    FormControlType = "radio"
	TextAreaFormControl FormControlType = "textarea"
	JsonFormControl     FormControlType = "json"
)
