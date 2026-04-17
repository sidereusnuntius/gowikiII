package e2e

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sidereusnuntius/gowiki/internal/model"
)

func (r *TestRig) fetchRemoteArticle(t *testing.T, id, expectedContent string) func(t *testing.T) {
	values := url.Values{
		"query": {id},
	}

	url := r.Wiki.Config.URL.JoinPath("search")
	url.RawQuery = values.Encode()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url.String(), nil)
	fatalErr(t, err)

	return func(t *testing.T) {
		_, _ = r.doRequest(t, req)

		req := model.ArticleRequest{
			IRI: id,
		}
		article, err := r.Wiki.ArticlesHandler.ArticleService.ArticleContent(t.Context(), &req)
		fatalErr(t, err)

		if diff := cmp.Diff(expectedContent, article.Content); len(diff) > 0 {
			t.Error(diff)
		}
	}
}

func (r *TestRig) editLocalArticle(t *testing.T, slug, host, content, summary string) func(t *testing.T) {
	var endpoint string
	if len(host) > 0 {
		endpoint = r.Wiki.Config.URL.JoinPath("a", host, slug, "edit").String()
	} else {
		endpoint = r.Wiki.Config.URL.JoinPath("a", slug, "edit").String()
	}

	var form bytes.Buffer
	writer := multipart.NewWriter(&form)
	writer.WriteField("content", content)
	writer.WriteField("summary", summary)
	writer.Close()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, endpoint, &form)
	fatalErr(t, err)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	return func(t *testing.T) {
		code, _ := r.doRequest(t, req)
		if code != http.StatusOK {
			t.Fatalf("unexpected code: %d", code)
		}

		articleReq := model.ArticleRequest{
			Host: host,
			Slug: slug,
		}
		article, err := r.Wiki.ArticlesHandler.ArticleService.ArticleContent(t.Context(), &articleReq)
		fatalErr(t, err)

		if article.Content != content {
			t.Error("saved article contains a different source")
			t.Logf("expected: \"%s\"", content)
			t.Logf("got: \"%s\"", article.Content)
		}
	}
}
