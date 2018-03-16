package jenkins

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/antchfx/xquery/xml"
)

type Jenkins struct {
	Client   *http.Client
	Server   string
	Username string
	Token    string
}

func NewJenkins(server, username, token string) *Jenkins {
	return &Jenkins{
		Client:   &http.Client{},
		Server:   server,
		Username: username,
		Token:    token,
	}
}

func (j *Jenkins) request(method, path, suffix string, body io.Reader) (*xmlquery.Node, error) {
	// Support paths already including the server part.
	if !strings.HasPrefix(path, j.Server) {
		path = j.Server + path
	}

	if len(suffix) > 0 {
		path += suffix
	}

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	// Credentials are needed on every requrest.
	req.SetBasicAuth(j.Username, j.Token)

	// When posting new content, the header is required.
	if method == "POST" && body != nil {
		req.Header.Add("Content-Type", "application/xml; charset=utf-8")
	}

	res, err := j.Client.Do(req)
	if err != nil {
		return nil, err
	}

	// Do hacky error handling.
	if res.StatusCode != 200 {
		bs, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(bs))
	}

	// Assume the content is xml and parse it.
	doc, err := xmlquery.Parse(res.Body)
	if err != nil {
		return nil, err
	}

	// remove declaration block <?xml ...
	// Jenkins encodes declaration block with single quotes, disliked by the parser.
	if doc != nil && doc.FirstChild != nil && doc.FirstChild.Type == xmlquery.DeclarationNode {
		doc.FirstChild = doc.FirstChild.NextSibling
		if doc.FirstChild != nil {
			doc.FirstChild.PrevSibling = nil
		}
	}

	return doc, nil
}

// Make a regular api request.
func (j *Jenkins) QueryApi(path string) (*xmlquery.Node, error) {
	return j.request("GET", path, "api/xml", nil)
}

// Request config.xml for job.
func (j *Jenkins) GetConfig(path string) (*xmlquery.Node, error) {
	return j.request("GET", path, "config.xml", nil)
}

// Update config.xml for job.
func (j *Jenkins) SendConfig(path string, config *xmlquery.Node) error {
	reader := getReaderFromNode(config)
	_, err := j.request("POST", path, "config.xml", reader)
	return err
}

// Request all projects recursively for a folder.
func (j *Jenkins) GetProjects(path string) ([]string, error) {
	doc, err := j.QueryApi(path)
	if err != nil {
		return nil, err
	}

	var projects []string

	for _, projectUrl := range xmlquery.Find(doc, "/folder/job[@_class='hudson.model.FreeStyleProject']/url") {
		projects = append(projects, projectUrl.InnerText())
	}

	for _, folderUrl := range xmlquery.Find(doc, "/folder/job[@_class='com.cloudbees.hudson.plugins.folder.Folder']/url") {
		childProjects, err := j.GetProjects(folderUrl.InnerText())
		if err != nil {
			return nil, err
		}

		projects = append(projects, childProjects...)
	}

	return projects, nil
}
