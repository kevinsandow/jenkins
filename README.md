# Jenkins API Client to update config.xml

Ever wanted to update a bunch of jenkins jobs all at once?

## Example: Setting Post-Build Action Notify via E-Mail:

```go
package main

import (
    "github.com/KevinBusse/jenkins"
    "github.com/antchfx/xquery/xml"
)

func main() {
    // Define jenkins apis configuration.
    j := jenkins.NewJenkins("https://...", "<username>", "<token>")

    // List all projects (recursively) for provided Folder.
    projects, err := j.GetProjects("job/Foo/job/Bar")  
    if err != nil {
        panic(err)
    }

    for _, project := range projects {
        // Request config.xml
        config, err := j.GetConfig(project)
        if err != nil {
            panic(err)
        }

        modifyConfig(config)

        // Save new config.xml
        err = j.SendConfig(project, config)
        if err != nil {
            panic(err)
        }
    }
}

func modifyConfig(config *xmlquery.Node) {
    // Find or create publisher node = post-build plugins.
    publishersNode := xmlquery.FindOne(config, "/project/publishers")
    if publishersNode == nil {
        projectNode := xmlquery.FindOne(config, "/project")
        if projectNode != nil {
            publishersNode = jenkins.AddElement(projectNode, "publishers")
        }
    }

    // Find or create new Mail Notification plugin.
    mailerNode := xmlquery.FindOne(publishersNode, "./hudson.tasks.Mailer")
    if mailerNode == nil {
        mailerNode = jenkins.AddElement(publishersNode, "hudson.tasks.Mailer")
    }
    jenkins.SetAttr(mailerNode, "plugin", "mailer@1.20")
    
    // Set new values for config.
    jenkins.SetElementTextNode(mailerNode, "recipients", "alerts@company.com")
    jenkins.SetElementTextNode(mailerNode, "dontNotifyEveryUnstableBuild", "false")
    jenkins.SetElementTextNode(mailerNode, "sendToIndividuals", "true")
}
```