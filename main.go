package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func main() {
	stackName := flag.String("s", "", "StackName")
	depth := flag.Int("d", 0, "Depth")
	flag.Parse()

	svc := cloudformation.New(getSession())
	current_b, err := getTemplateBody(svc, *stackName, 0)
	if err != nil {
		log.Fatal(err)
	}

	prev_b, err := getTemplateBody(svc, *stackName, *depth)
	if err != nil {
		log.Fatal(err)
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(*current_b, *prev_b, false)
	fmt.Println(dmp.DiffPrettyText(dmp.DiffCleanupSemantic(diffs)))
}

// for using assume role
func getSession() *session.Session {
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}
	return session.Must(session.NewSessionWithOptions(opts))
}

func getChangeSetIDWithDepth(svc *cloudformation.CloudFormation, stackName string, depth int) (*string, error) {
	o, err := svc.ListChangeSets(&cloudformation.ListChangeSetsInput{
		StackName: &stackName,
	})
	if err != nil {
		return nil, err
	}
	if len(o.Summaries)-depth < 0 {
		return nil, fmt.Errorf("index out of range, length:%d, index:%d", len(o.Summaries), len(o.Summaries)-depth)
	}
	return o.Summaries[len(o.Summaries)-depth].ChangeSetId, nil
}

func getTemplateBody(svc *cloudformation.CloudFormation, stackName string, depth int) (*string, error) {
	var input cloudformation.GetTemplateInput
	if depth > 0 {
		changeSetID, err := getChangeSetIDWithDepth(svc, stackName, depth)
		if err != nil {
			return nil, err
		}
		input = cloudformation.GetTemplateInput{
			StackName:     &stackName,
			ChangeSetName: changeSetID,
		}
		fmt.Println(changeSetID)
	} else {
		input = cloudformation.GetTemplateInput{
			StackName: &stackName,
		}
	}

	o, err := svc.GetTemplate(&input)
	if err != nil {
		return nil, err
	}

	return o.TemplateBody, nil
}
