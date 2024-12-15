package extractors

import "testing"

func TestGetUrl(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{name: "t1", url: "https://streamff.com/v/e70b90d8", want: "https://ffedge.streamff.com/uploads/e70b90d8.mp4"},
		{name: "t2", url: "https://streamff.com/v/e70b90d8/", want: "https://ffedge.streamff.com/uploads/e70b90d8.mp4"},
		{name: "t3", url: "https://streamff.com/v/e70b90d8/test", want: "https://ffedge.streamff.com/uploads/e70b90d8.mp4"},
		{name: "t4", url: "https://streamff.com/v/e70b90d8?test", want: "https://ffedge.streamff.com/uploads/e70b90d8.mp4"},
		{name: "t5", url: "https://www.streamff.com/v/e70b90d8", want: "https://ffedge.streamff.com/uploads/e70b90d8.mp4"},
		{name: "t6", url: "streamff.com/v/e70b90d8?test", want: "https://ffedge.streamff.com/uploads/e70b90d8.mp4"},
		{name: "t7", url: "www.streamff.com/v/e70b90d8?test", want: "https://ffedge.streamff.com/uploads/e70b90d8.mp4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUrl(tt.url); got != tt.want {
				t.Errorf("GetUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
