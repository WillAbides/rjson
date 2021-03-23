package benchmarks

import (
	"compress/gzip"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	benchInt  int
	benchBool bool
)

type benchSample struct {
	name string
	data []byte
}

func benchSamples(t testing.TB) []benchSample {
	return []benchSample{
		{
			name: "github user",
			data: []byte(exampleGithubUser),
		},
		{
			name: "large object",
			data: getTestdataJSONGz(t, "citm_catalog.json"),
		},
		{
			name: "string",
			data: []byte(`"this is a simple string"`),
		},
		{
			name: "integer value",
			data: []byte(`1234567`),
		},
		{
			name: "float value",
			data: []byte(`12.34567`),
		},
		{
			name: "null",
			data: []byte(`null`),
		},
	}
}

func gunzipTestJSON(t testing.TB, filename string) string {
	t.Helper()
	targetDir := filepath.Join("..", "testdata", "tmp")
	err := os.MkdirAll(targetDir, 0o700)
	require.NoError(t, err)
	target := filepath.Join(targetDir, filename)
	if fileExists(t, target) {
		return target
	}
	src := filepath.Join("..", "testdata", filename+".gz")
	f, err := os.Open(src)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, f.Close())
	}()
	gz, err := gzip.NewReader(f)
	require.NoError(t, err)
	buf, err := ioutil.ReadAll(gz)
	require.NoError(t, err)
	err = ioutil.WriteFile(target, buf, 0o600)
	require.NoError(t, err)
	return target
}

func getTestdataJSONGz(t testing.TB, path string) []byte {
	t.Helper()
	filename := gunzipTestJSON(t, path)
	got, err := ioutil.ReadFile(filename)
	require.NoError(t, err)
	return got
}

func fileExists(t testing.TB, filename string) bool {
	t.Helper()
	_, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	require.NoError(t, err)
	return true
}

var exampleGithubUser = `{
 "avatar_url": "https://avatars.githubusercontent.com/u/583231?v=4",
 "bio": null,
 "blog": "https://github.blog",
 "company": "@github",
 "created_at": "2011-01-25T18:44:36Z",
 "email": "octocat@github.com",
 "events_url": "https://api.github.com/users/octocat/events{/privacy}",
 "followers": 3599,
 "followers_url": "https://api.github.com/users/octocat/followers",
 "following": 9,
 "following_url": "https://api.github.com/users/octocat/following{/other_user}",
 "gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
 "gravatar_id": "",
 "hireable": null,
 "html_url": "https://github.com/octocat",
 "id": 583231,
 "location": "San Francisco",
 "login": "octocat",
 "name": "The Octocat",
 "node_id": "MDQ6VXNlcjU4MzIzMQ==",
 "organizations_url": "https://api.github.com/users/octocat/orgs",
 "public_gists": 8,
 "public_repos": 8,
 "received_events_url": "https://api.github.com/users/octocat/received_events",
 "repos_url": "https://api.github.com/users/octocat/repos",
 "site_admin": false,
 "starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
 "subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
 "twitter_username": null,
 "type": "User",
 "updated_at": "2021-03-22T14:27:47Z",
 "url": "https://api.github.com/users/octocat"
}`
