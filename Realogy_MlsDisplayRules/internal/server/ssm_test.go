package server

import "testing"

//TODO
func Test_interpolate(t *testing.T) {
	type args struct {
		c    Context
		text string
	}
	var tests []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := interpolate(tt.args.c, tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("interpolate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("interpolate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
