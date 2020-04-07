package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	googleoauth "golang.org/x/oauth2/google"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var defaultClientScopes = []string{
	"https://www.googleapis.com/auth/cloud-platform",
}

func Provisioner() terraform.ResourceProvisioner {
	return &schema.Provisioner{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Description: "REST URL to call",
				Required:    true,
				ValidateFunc: func(value interface{}, name string) ([]string, []error) {
					val, ok := value.(string)
					if !ok {
						return nil, []error{errors.New("invalid type for url, expected string")}
					}
					_, err := url.Parse(val)
					if err != nil {
						return nil, []error{err}
					}
					return nil, nil
				},
			},
			"method": {
				Type:        schema.TypeString,
				Description: "REST Method (GET,PUSH,PUT,PATCH,DELETE)",
				Required:    true,
				ValidateFunc: func(value interface{}, name string) ([]string, []error) {
					val, ok := value.(string)
					if !ok {
						return nil, []error{errors.New("invalid type for method, expected string")}
					}
					switch strings.ToUpper(val) {
					case "GET":
					case "POST":
					case "PATCH":
					case "DELETE":
					default:
						return nil, []error{
							errors.New("invalid http method: " + val),
						}
					}
					return nil, nil
				},
			},
			"json": {
				Type:        schema.TypeString,
				Description: "Json content of the request",
				Optional:    true,
			},
			"access_token": {
				Type:        schema.TypeString,
				Description: "Gcp access token",
				Optional:    true,
			},
		},
		ApplyFunc: func(ctx context.Context) error {
			var err error
			var ok bool
			var output terraform.UIOutput
			var url_string string
			var method string
			var json string
			output = ctx.Value(schema.ProvOutputKey).(terraform.UIOutput)
			var config *schema.ResourceData
			config = ctx.Value(schema.ProvConfigDataKey).(*schema.ResourceData)
			url_string, ok = config.Get("url").(string)
			if !ok {
				return errors.New("invalid url")
			}
			url_value, err := url.Parse(url_string)
			if err != nil {
				return err
			}
			method, ok = config.Get("method").(string)
			if !ok {
				return errors.New("invalid method")
			}
			json_val := config.Get("json")
			if json_val != nil {
				json, ok = json_val.(string)
				if !ok {
					return errors.New("invalid json")
				}
			} else {
				json = ""
			}
			token_data, ok := config.Get("token").(string)
			var token_source oauth2.TokenSource
			if ok && token_data != "" {
				token := &oauth2.Token{AccessToken: token_data}
				token_source = oauth2.StaticTokenSource(token)
			} else {
				token_source, err = googleoauth.DefaultTokenSource(ctx, defaultClientScopes...)
				if err != nil {
					return err
				}
			}
			client := oauth2.NewClient(ctx, token_source)
			var req *http.Request
			if json != "" {
				reader := ioutil.NopCloser(bytes.NewReader([]byte(json)))
				req = &http.Request{
					Method: method,
					URL:    url_value,
					Body:   reader,
				}
			} else {
				req = &http.Request{
					Method: method,
					URL:    url_value,
				}
			}
			output.Output(fmt.Sprintf("Executing %s on %s", method, url_value.String()))
			res, err := client.Do(req)
			if err != nil {
				return err
			}
			if res.StatusCode > 299 || res.StatusCode < 200 {
				// Return an error if status code is not in the 200 range
				return fmt.Errorf("http request returned status %d", res.StatusCode)
			} else {
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					return err
				}
				output.Output(string(body))
			}
			return nil
		},
	}
}
