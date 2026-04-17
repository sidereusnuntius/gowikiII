package e2e

import (
	"testing"
)

func Test(t *testing.T) {
	rig1 := Start(t, "http://localhost:8082")
	rig2 := Start(t, "http://localhost:8083")

	t.Run("sign up on first instance", rig1.createUser(t, "johndoe", "strongpass123", "jdoe@example.com"))
	t.Run("sign up on second instance", rig2.createUser(t, "alice", "coxinha123", "theonealice@aol.com"))

	slug := "hello_world"
	content := "\"Hello world\" is a program that programmers traditionally write when learning to program."
	t.Run(
		"create article on instance 1",
		rig1.editLocalArticle(t, slug, "", content, "created new article"),
	)

	articleURL := rig1.Wiki.Config.URL.JoinPath("a", "hello_world").String()
	t.Run("fetch article from instance 2", rig2.fetchRemoteArticle(t, articleURL, content))

	rig1.Close(t)
	rig2.Close(t)
	t.Log("shutdown server! waiting")
	wg.Wait()
	t.Log("finished")
}
