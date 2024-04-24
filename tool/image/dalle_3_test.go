package tools

import (
	"fmt"
	"reflect"
	"testing"
)

func TestDalleDraw_ArgsType(t *testing.T) {
	type fields struct {
		opts DalleOptions
	}
	tests := []struct {
		name   string
		fields fields
		want   reflect.Type
	}{
		{
			name: "test",
			fields: fields{
				opts: DalleOptions{},
			},
			want: reflect.TypeOf(&DalleDrawRequest{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := DalleDraw{
				opts: tt.fields.opts,
			}
			got := c.ArgsType()
			fmt.Println(got, tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				fmt.Println(got, tt.want)
				t.Errorf("ArgsType() = %v, want %v", got, tt.want)
			}

		})
	}
}
