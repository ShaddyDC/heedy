package schema

import (
	"testing"
)

func TestSchema(t *testing.T) {
	s_string, err := NewSchema(`{"type": "string"}`)
	if err != nil {
		t.Errorf("Failed to create schema : %s", err)
		return
	}
	s_float, err := NewSchema(`{"type": "number"}`)
	if err != nil {
		t.Errorf("Failed to create schema : %s", err)
		return
	}
	s_obj, err := NewSchema(`{"type": "object", "properties": {"lat": {"type": "number"},"msg": {"type": "string"}}}`)
	if err != nil {
		t.Errorf("Failed to create schema : %s", err)
		return
	}

	v_string := "Hello"
	v_float := 3.14
	v_obj := map[string]interface{}{"lat": 88.32, "msg": "hi"}
	v_bobj := map[string]interface{}{"lat": "88.32", "msg": "hi"}

	if !s_string.IsValid(v_string) || !s_float.IsValid(v_float) || !s_obj.IsValid(v_obj) {
		t.Errorf("Validation failed")
		return
	}
	if s_obj.IsValid(v_bobj) {
		t.Errorf("Validation wrong")
		return
	}

	var x interface{}

	val, err := s_string.Marshal(v_string)
	if err != nil {
		t.Errorf("Marshal failed")
		return
	}

	if err = s_string.Unmarshal(val, &x); err != nil {
		t.Errorf("unmarshal failed")
		return
	}
	if v, ok := x.(string); !ok || v != v_string {
		t.Errorf("Crap: %v, %v", ok, v)
		return
	}

	val, err = s_float.Marshal(v_float)
	if err != nil {
		t.Errorf("Marshal failed")
		return
	}

	if err = s_float.Unmarshal(val, &x); err != nil {
		t.Errorf("unmarshal failed")
		return
	}
	if v, ok := x.(float64); !ok || v != v_float {
		t.Errorf("Crap: %v, %v", ok, v)
		return
	}

	val, err = s_obj.Marshal(v_obj)
	if err != nil {
		t.Errorf("Marshal failed")
		return
	}

	if err = s_obj.Unmarshal(val, &x); err != nil {
		t.Errorf("unmarshal failed")
		return
	}
	if v, ok := x.(map[string]interface{}); !ok || v["lat"].(float64) != 88.32 || v["msg"].(string) != "hi" {
		t.Errorf("Crap: %v, %v", ok, v)
		return
	}
}
