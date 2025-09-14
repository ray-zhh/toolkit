package httputils

import "testing"

func Test_GetFileNameFromUrl(t *testing.T) {
	url := "https://www.gfxcamp.com/wp-content/uploads/2014/05/VideoHive-Web-Promotion-Guide-.jpg"
	fileName, err := GetFileNameFromUrl(url)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(fileName)
}
