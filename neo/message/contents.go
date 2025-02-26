package message

import (
	"strings"

	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
)

const (
	// ContentStatusPending the content status pending
	ContentStatusPending = iota
	// ContentStatusDone the content status done
	ContentStatusDone
	// ContentStatusError the content status error
	ContentStatusError
)

var tokens = map[string][2]string{
	"think": {"<think>", "</think>"},
	"tool":  {"<tool>", "</tool>"},
}

// Contents the contents
type Contents struct {
	Current int    `json:"current"` // the current content index
	Data    []Data `json:"data"`    // the data
	token   string // the current token
	id      string // the id of the contents
}

// Data the data of the content
type Data struct {
	Type  string                 `json:"type"`  // text, function, error, think, tool
	ID    string                 `json:"id"`    // the id of the content
	Bytes []byte                 `json:"bytes"` // the content bytes
	Props map[string]interface{} `json:"props"` // the props
}

// NewContents create a new contents
func NewContents() *Contents {
	return &Contents{
		Current: -1,
		Data:    []Data{},
	}
}

// ScanTokens scan the tokens
func (c *Contents) ScanTokens(currentID string, cb func(token string, id string, begin bool, text string, tails string)) {

	text := strings.TrimSpace(c.Text())

	// check the end of the token
	if c.token != "" {
		token := tokens[c.token]

		// Check the end of the token
		if index := strings.Index(text, token[1]); index >= 0 {
			tails := ""
			if index > 0 {
				tails = text[index+len(token[1]):]
				text = text[:index+len(token[1])]
			}
			c.UpdateType(c.token, map[string]interface{}{"text": text}, c.id)
			c.NewText([]byte(tails), c.id) // Create new text with the tails
			cb(c.token, c.id, false, text, tails)
			c.ClearToken() // clear the token
			return
		}

		// call the callback for the begin of the token
		cb(c.token, c.id, true, text, "")
		return
	}

	// scan the begin of the token
	for name, token := range tokens {
		if index := strings.Index(text, token[0]); index >= 0 {
			c.token = name
			c.id = currentID
			if c.id == "" {
				c.id = uuid.New().String()
			}
			cb(name, c.id, true, text, "") // call the callback
		}
	}
}

// ClearToken clear the token
func (c *Contents) ClearToken() {
	c.token = ""
}

// RemoveLastEmpty remove the last empty data
func (c *Contents) RemoveLastEmpty() {
	if c.Current == -1 {
		return
	}

	// Remove the last empty data
	if len(c.Data[c.Current].Bytes) == 0 && c.Data[c.Current].Type == "text" {
		c.Data = c.Data[:c.Current]
		c.Current--
	}
}

// NewText create a new text data and append to the contents
func (c *Contents) NewText(bytes []byte, id ...string) *Contents {

	data := Data{Type: "text", Bytes: bytes}
	if len(id) > 0 && id[0] != "" {
		data.ID = id[0]
	}
	c.Data = append(c.Data, data)
	c.Current++
	return c
}

// NewType create a new type data and append to the contents
func (c *Contents) NewType(typ string, props map[string]interface{}, id ...string) *Contents {

	data := Data{
		Type:  typ,
		Props: props,
	}
	if len(id) > 0 && id[0] != "" {
		data.ID = id[0]
	}
	c.Data = append(c.Data, data)
	c.Current++
	return c
}

// UpdateType update the type of the current content
func (c *Contents) UpdateType(typ string, props map[string]interface{}, id ...string) *Contents {
	if c.Current == -1 {
		c.NewType(typ, props, id...)
		return c
	}

	if len(id) > 0 && id[0] != "" {
		c.Data[c.Current].ID = id[0]
	}
	c.Data[c.Current].Type = typ
	// c.Data[c.Current].Props = props
	if c.Data[c.Current].Props == nil  {
		c.Data[c.Current].Props = props
	}else{
		for key, value := range props {
			c.Data[c.Current].Props[key] = value
		}
	}
	
	return c
}

// NewError create a new error data and append to the contents
func (c *Contents) NewError(err []byte) *Contents {
	c.Data = append(c.Data, Data{
		Type:  "error",
		Bytes: err,
	})
	c.Current++
	return c
}

// AppendText append the text to the current content
func (c *Contents) AppendText(bytes []byte, id ...string) *Contents {
	if c.Current == -1 {
		c.NewText(bytes, id...)
		return c
	}

	if len(id) > 0 && id[0] != "" {
		c.Data[c.Current].ID = id[0]
	}
	c.Data[c.Current].Bytes = append(c.Data[c.Current].Bytes, bytes...)
	return c
}

// AppendError append the error to the current content
func (c *Contents) AppendError(err []byte) *Contents {
	if c.Current == -1 {
		c.NewError(err)
		return c
	}
	c.Data[c.Current].Bytes = append(c.Data[c.Current].Bytes, err...)
	return c
}

// JSON returns the json representation
func (c *Contents) JSON() string {
	raw, _ := jsoniter.MarshalToString(c.Data)
	return raw
}

// Text returns the text of the current content
func (c *Contents) Text() string {
	if c.Current == -1 {
		return ""
	}
	return string(c.Data[c.Current].Bytes)
}

// CurrentType returns the type of the current content
func (c *Contents) CurrentType() string {
	if c.Current == -1 {
		return ""
	}
	return c.Data[c.Current].Type
}

// Map returns the map representation
func (data *Data) Map() (map[string]interface{}, error) {
	v := map[string]interface{}{"type": data.Type}

	if data.ID != "" {
		v["id"] = data.ID
	}

	if data.Bytes != nil && data.Type == "text" {
		v["text"] = string(data.Bytes)
	}

	if data.Props != nil && data.Type != "text" {
		v["props"] = data.Props
	}
	return v, nil
}

// MarshalJSON returns the json representation
func (data *Data) MarshalJSON() ([]byte, error) {

	v := map[string]interface{}{"type": data.Type}

	if data.ID != "" {
		v["id"] = data.ID
	}

	if data.Bytes != nil && data.Type == "text" {
		v["text"] = string(data.Bytes)
	}

	if data.Props != nil && data.Type != "text" {
		v["props"] = data.Props
	}

	return jsoniter.Marshal(v)
}
func (data *Data) UnmarshalJSON(input []byte) error {
    origin := map[string]interface{}{}
    err := jsoniter.Unmarshal(input, &origin)
    if err != nil {
        return err
    }

    // Initialize empty maps/slices
    newData := Data{
        Props: make(map[string]interface{}),
    }

    // Safe type assertions with default values
    if id, ok := origin["id"].(string); ok {
        newData.ID = id
    }
    
    if typ, ok := origin["type"].(string); ok {
        newData.Type = typ
    }
    
    if text, ok := origin["text"].(string); ok {
        newData.Bytes = []byte(text)
    }
    
    if props, ok := origin["props"].(map[string]interface{}); ok {
        newData.Props = props
    }

    *data = newData
    return nil
}
