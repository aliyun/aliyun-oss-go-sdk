package oss

import "testing"
import "fmt"

func TestGenerateSign(t *testing.T) {

	compositions1 := &Compositions{
		Method:          "GET",
		Date:            "1493989264",
		AccessKeySecret: "6edKtOCgTV5XDCrVTqHgHA9pgtWRi8",
	}
	compositions2 := &Compositions{
		Method:          "GET",
		Date:            "Thu, 04 May 2017 10:50:23 GMT",
		AccessKeySecret: "we32tOCgTV5XDCretqHgHA9pdwrRi8",
	}

	type args struct {
		flag                  int
		canonicalizedResource string
		compositions          *Compositions
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				flag: 1,
				canonicalizedResource: "/test/0000000001/0000000001.20151018.3b3e9730-7541-11e5-9480-2f80e994f28c.json",
				compositions:          compositions1,
			},
			want:    "wfJvYphml7qGtYLO9Up4UikxW0E%3D",
			wantErr: false,
		}, {
			name: "2",
			args: args{
				flag: 2,
				canonicalizedResource: "/test/0000000001/0000000001.20151018.3b3e9730-7541-11e5-9480-2f80e994f28c.json",
				compositions:          compositions2,
			},
			want:    "OSS LTAINwY5Hri5wwQL:ClnfeWr0AFc6MyduDbyS4E1TZFM=",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateSign(tt.args.flag, tt.args.canonicalizedResource, tt.args.compositions)
			if err != nil {
				t.Errorf("GenerateSign() error = %v, wantErr %v", err, tt.wantErr)
			}
			fmt.Println(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateSign() = %v, want %v", got, tt.want)
			}
			t.Log("test case ok!")
		})
	}
}
