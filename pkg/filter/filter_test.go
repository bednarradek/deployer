package filter

import (
	"testing"
)

func TestFileFilter_Contain(t *testing.T) {
	type fields struct {
		ignoreList []string
	}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Test 1",
			fields: fields{
				ignoreList: []string{".DS_Store"},
			},
			args: args{
				path: "/.DS_Store",
			},
			want: true,
		},
		{
			name: "Test 2",
			fields: fields{
				ignoreList: []string{".DS_Store"},
			},
			args: args{
				path: "/app/.DS_Store",
			},
			want: true,
		},
		{
			name: "Test 3",
			fields: fields{
				ignoreList: []string{".DS_Store"},
			},
			args: args{
				path: "/app/bootstrap.php",
			},
			want: false,
		},
		{
			name: "Test 4",
			fields: fields{
				ignoreList: []string{"/vendor"},
			},
			args: args{
				path: "/vendor",
			},
			want: true,
		},
		{
			name: "Test 5",
			fields: fields{
				ignoreList: []string{"/vendor"},
			},
			args: args{
				path: "vendor",
			},
			want: false,
		},
		{
			name: "Test 6",
			fields: fields{
				ignoreList: []string{"/package.*"},
			},
			args: args{
				path: "/package.json",
			},
			want: true,
		},
		{
			name: "Test 7",
			fields: fields{
				ignoreList: []string{"/package.*"},
			},
			args: args{
				path: "/package-lock.json",
			},
			want: true,
		},
		{
			name: "Test 8",
			fields: fields{
				ignoreList: []string{"^/src"},
			},
			args: args{
				path: "/src",
			},
			want: true,
		},
		{
			name: "Test 9",
			fields: fields{
				ignoreList: []string{"^/src"},
			},
			args: args{
				path: "/vendor/src",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileSystemFilter{
				ignoreList: tt.fields.ignoreList,
			}
			if got := f.Contain(tt.args.path); got != tt.want {
				t.Errorf("Contain() = %v, want %v", got, tt.want)
			}
		})
	}
}
