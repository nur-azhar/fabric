package common

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"io/ioutil"
	"net/http"
)

type Attachment struct {
	Type    *string `json:"type,omitempty"`
	Path    *string `json:"path,omitempty"`
	URL     *string `json:"url,omitempty"`
	Content []byte  `json:"content,omitempty"`
	ID      *string `json:"id,omitempty"`
}

func (a *Attachment) GetId() (ret string, err error) {
	if a.ID == nil {
		var hash string
		if a.Content != nil {
			hash = fmt.Sprintf("%x", sha256.Sum256(a.Content))
		} else if a.Path != nil {
			var content []byte
			if content, err = ioutil.ReadFile(*a.Path); err != nil {
				return
			}
			hash = fmt.Sprintf("%x", sha256.Sum256(content))
		} else if a.URL != nil {
			data := map[string]string{"url": *a.URL}
			var jsonData []byte
			if jsonData, err = json.Marshal(data); err != nil {
				return
			}
			hash = fmt.Sprintf("%x", sha256.Sum256(jsonData))
		}
		a.ID = &hash
	}
	ret = *a.ID
	return
}

func (a *Attachment) ResolveType() (ret string, err error) {
	if a.Type != nil {
		ret = *a.Type
		return
	}
	if a.Path != nil {
		var mime *mimetype.MIME
		if mime, err = mimetype.DetectFile(*a.Path); err != nil {
			return
		}
		ret = mime.String()
		return
	}
	if a.URL != nil {
		var resp *http.Response
		if resp, err = http.Head(*a.URL); err != nil {
			return
		}
		defer resp.Body.Close()
		ret = resp.Header.Get("Content-Type")
		return
	}
	if a.Content != nil {
		ret = mimetype.Detect(a.Content).String()
		return
	}
	err = fmt.Errorf("attachment has no type and no content to derive it from")
	return
}

func (a *Attachment) ContentBytes() (ret []byte, err error) {
	if a.Content != nil {
		ret = a.Content
		return
	}
	if a.Path != nil {
		if ret, err = ioutil.ReadFile(*a.Path); err != nil {
			return
		}
		return
	}
	if a.URL != nil {
		var resp *http.Response
		if resp, err = http.Get(*a.URL); err != nil {
			return
		}
		defer resp.Body.Close()
		if ret, err = ioutil.ReadAll(resp.Body); err != nil {
			return
		}
		return
	}
	err = fmt.Errorf("no content available")
	return
}

func (a *Attachment) Base64Content() (ret string, err error) {
	var content []byte
	if content, err = a.ContentBytes(); err != nil {
		return
	}
	ret = base64.StdEncoding.EncodeToString(content)
	return
}

func FromRow(row map[string]interface{}) (ret *Attachment, err error) {
	attachment := &Attachment{}
	if id, ok := row["id"].(string); ok {
		attachment.ID = &id
	}
	if typ, ok := row["type"].(string); ok {
		attachment.Type = &typ
	}
	if path, ok := row["path"].(string); ok {
		attachment.Path = &path
	}
	if url, ok := row["url"].(string); ok {
		attachment.URL = &url
	}
	if content, ok := row["content"].([]byte); ok {
		attachment.Content = content
	}
	ret = attachment
	return
}
