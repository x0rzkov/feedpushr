package main

import (
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/ChimeraCoder/anaconda"
	"github.com/ncarlier/feedpushr/v2/pkg/expr"
	"github.com/ncarlier/feedpushr/v2/pkg/format"
	"github.com/ncarlier/feedpushr/v2/pkg/model"
)

var spec = model.Spec{
	Name: "twitter",
	Desc: "Send new articles to a Twitter timeline.",
	PropsSpec: []model.PropSpec{
		{
			Name: "consumerKey",
			Desc: "Consumer key",
			Type: model.Text,
		},
		{
			Name: "consumerSecret",
			Desc: "Consumer secret",
			Type: model.Password,
		},
		{
			Name: "accessToken",
			Desc: "Access token",
			Type: model.Text,
		},
		{
			Name: "accessTokenSecret",
			Desc: "Access token secret",
			Type: model.Password,
		},
		{
			Name: "format",
			Desc: "Tweet format (default: `{{.Title}}\\n{{.Link}}`)",
			Type: model.Textarea,
		},
	},
}

// TwitterOutputPlugin is the Twitter output plugin
type TwitterOutputPlugin struct{}

// Spec returns plugin spec
func (p *TwitterOutputPlugin) Spec() model.Spec {
	return spec
}

// Build creates Twitter output provider instance
func (p *TwitterOutputPlugin) Build(output *model.OutputDef) (model.OutputProvider, error) {
	condition, err := expr.NewConditionalExpression(output.Condition)
	if err != nil {
		return nil, err
	}
	// Default format
	if frmt, ok := output.Props["format"]; !ok || frmt == "" {
		output.Props["format"] = "{{.Title}}\n{{.Link}}"
	}
	formatter, err := format.NewOutputFormatter(output)
	if err != nil {
		return nil, err
	}
	consumerKey := output.Props.Get("consumerKey")
	if consumerKey == "" {
		return nil, fmt.Errorf("missing consumer key property")
	}
	consumerSecret := output.Props.Get("consumerSecret")
	if consumerSecret == "" {
		return nil, fmt.Errorf("missing consumer secret property")
	}
	accessToken := output.Props.Get("accessToken")
	if accessToken == "" {
		return nil, fmt.Errorf("missing access token property")
	}
	accessTokenSecret := output.Props.Get("accessTokenSecret")
	if accessTokenSecret == "" {
		return nil, fmt.Errorf("missing access token secret property")
	}
	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	api := anaconda.NewTwitterApi(accessToken, accessTokenSecret)

	return &TwitterOutputProvider{
		id:             output.ID,
		alias:          output.Alias,
		spec:           spec,
		condition:      condition,
		formatter:      formatter,
		enabled:        output.Enabled,
		api:            api,
		consumerKey:    consumerKey,
		consumerSecret: consumerSecret,
	}, nil
}

// TwitterOutputProvider output provider to send articles to Twitter
type TwitterOutputProvider struct {
	id             int
	alias          string
	spec           model.Spec
	condition      *expr.ConditionalExpression
	formatter      format.Formatter
	enabled        bool
	nbError        uint64
	nbSuccess      uint64
	consumerKey    string
	consumerSecret string
	api            *anaconda.TwitterApi
}

// Send sent an article as Tweet to a Twitter timeline
func (op *TwitterOutputProvider) Send(article *model.Article) error {
	if !op.enabled || !op.condition.Match(article) {
		// Ignore if disabled or if the article doesn't match the condition
		return nil
	}
	b, err := op.formatter.Format(article)
	if err != nil {
		atomic.AddUint64(&op.nbError, 1)
		return err
	}
	tweet := truncate(b.String(), 280)
	v := url.Values{}
	_, err = op.api.PostTweet(tweet, v)
	if err != nil {
		// Ignore error due to duplicate status
		if strings.Contains(err.Error(), "\"code\":187") {
			return nil
		}
		atomic.AddUint64(&op.nbError, 1)
	} else {
		atomic.AddUint64(&op.nbSuccess, 1)
	}
	return err
}

// GetDef return filter definition
func (op *TwitterOutputProvider) GetDef() model.OutputDef {
	result := model.OutputDef{
		ID:        op.id,
		Alias:     op.alias,
		Spec:      op.spec,
		Condition: op.condition.String(),
		Enabled:   op.enabled,
	}
	result.Props = map[string]interface{}{
		"consumerKey":       op.consumerKey,
		"consumerSecret":    op.consumerSecret,
		"accessToken":       op.api.Credentials.Token,
		"accessTokenSecret": op.api.Credentials.Secret,
		"nbError":           op.nbError,
		"nbSuccess":         op.nbSuccess,
		"format":            op.formatter.Value(),
	}
	return result
}

// GetPluginSpec returns plugin spec
func GetPluginSpec() model.PluginSpec {
	return model.PluginSpec{
		Spec: spec,
		Type: model.OUTPUT_PLUGIN,
	}
}

// GetOutputPlugin returns output plugin
func GetOutputPlugin() (op model.OutputPlugin, err error) {
	return &TwitterOutputPlugin{}, nil
}
