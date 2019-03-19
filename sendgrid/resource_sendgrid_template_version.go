package sendgrid

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	homedir "github.com/mitchellh/go-homedir"
	sendgrid_client "github.com/syoya/terraform-client-sendgrid"
)

func resourceSendgridTemplateVersion() *schema.Resource {
	return &schema.Resource{
		Create: resourceSendgridTemplateVersionCreate,
		Read:   resourceSendgridTemplateVersionRead,
		Update: resourceSendgridTemplateVersionUpdate,
		Delete: resourceSendgridTemplateVersionDelete,
		Exists: resourceSendgridTemplateVersionExists,
		Importer: &schema.ResourceImporter{
			State: resourceSendgridTemplateVersionImport,
		},
		CustomizeDiff: resourceSendgridTemplateVersionCustomizeDiff,

		Schema: map[string]*schema.Schema{
			"template_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"subject": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"html_content_file": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"plain_content_file": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"html_content_hash": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"plain_content_hash": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"active": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func buildTemplateVersionStruct(d *schema.ResourceData) (*sendgrid_client.TemplateVersion, error) {
	htmlContent, err := loadFileContent(d.Get("html_content_file").(string))
	if err != nil {
		return nil, err
	}

	plainContent, err := loadFileContent(d.Get("plain_content_file").(string))
	if err != nil {
		return nil, err
	}

	active := 0
	if d.Get("active").(bool) {
		active = 1
	}

	m := sendgrid_client.TemplateVersion{
		TemplateId:   d.Get("template_id").(string),
		Name:         d.Get("name").(string),
		Subject:      d.Get("subject").(string),
		HtmlContent:  string(htmlContent),
		PlainContent: string(plainContent),
		Active:       active,
	}

	return &m, nil
}

func resourceSendgridTemplateVersionExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	client := meta.(*sendgrid_client.Client)

	fmt.Println("Exist template_version")
	if _, err := client.GetTemplateVersion(d.Get("template_id").(string), d.Id()); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, fmt.Errorf("error checking template_version: %s", err.Error())
	}

	return true, nil
}

func resourceSendgridTemplateVersionCreate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*sendgrid_client.Client)

	m, err := buildTemplateVersionStruct(d)
	if err != nil {
		return err
	}
	fmt.Println("Create template_version1")
	m, err = client.CreateTemplateVersion(m)
	if err != nil {
		return fmt.Errorf("error creating template_version: %s", err.Error())
	}
	fmt.Println("Create template_version2")
	d.SetId(m.Id)
	d.Set("html_content_hash", getHash(m.HtmlContent))
	d.Set("plain_content_hash", getHash(m.PlainContent))

	return nil
}

func resourceSendgridTemplateVersionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*sendgrid_client.Client)

	fmt.Println("Read template_version")
	m, err := client.GetTemplateVersion(d.Get("template_id").(string), d.Id())
	if err != nil {
		return fmt.Errorf("error reading template_version: %s", err.Error())
	}

	remoteHTMLHash := getHash(string(m.HtmlContent))
	remotePlainHash := getHash(string(m.PlainContent))

	stateHTMLHash := d.Get("html_content_hash")
	statePlainHash := d.Get("plain_content_hash")

	fmt.Printf("[DEBUG] TemplateVersion: %s ----- %s", remoteHTMLHash, stateHTMLHash)
	fmt.Printf("[DEBUG] TemplateVersion: %s ----- %s", remotePlainHash, statePlainHash)

	d.Set("html_content_hash", remoteHTMLHash)
	d.Set("plain_content_hash", remotePlainHash)
	return nil
}

func resourceSendgridTemplateVersionUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*sendgrid_client.Client)

	m, err := buildTemplateVersionStruct(d)
	if err != nil {
		return err
	}
	fmt.Println("Update template_version")
	if err := client.UpdateTemplateVersion(d.Id(), m); err != nil {
		return fmt.Errorf("error updating TemplateVersion: %s", err.Error())
	}

	d.Set("html_content_hash", getHash(m.HtmlContent))
	d.Set("plain_content_hash", getHash(m.PlainContent))

	return resourceSendgridTemplateVersionRead(d, meta)
}

func resourceSendgridTemplateVersionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*sendgrid_client.Client)

	fmt.Println("Delete template_version")
	if err := client.DeleteTemplateVersion(d.Get("template_id").(string), d.Id()); err != nil {
		return fmt.Errorf("error deleting TemplateVersion: %s", err.Error())
	}

	d.Set("html_content_hash", "")
	d.Set("plain_content_hash", "")

	return nil
}

func resourceSendgridTemplateVersionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	fmt.Println("Import template_version")
	if err := resourceSendgridTemplateVersionRead(d, meta); err != nil {
		return nil, fmt.Errorf("error importing TemplateVersion: %s", err.Error())
	}
	return []*schema.ResourceData{d}, nil
}

func resourceSendgridTemplateVersionCustomizeDiff(d *schema.ResourceDiff, meta interface{}) error {
	stateHTMLHash := d.Get("html_content_hash").(string)
	htmlContent, err := loadFileContent(d.Get("html_content_file").(string))
	if err != nil {
		return err
	}
	localHTMLHash := getHash(string(htmlContent))
	if stateHTMLHash != localHTMLHash {
		if err := d.SetNew("html_content_hash", localHTMLHash); err != nil {
			return err
		}
	}

	statePlainHash := d.Get("plain_content_hash").(string)
	plainContent, err := loadFileContent(d.Get("plain_content_file").(string))
	if err != nil {
		return err
	}
	localPlainHash := getHash(string(plainContent))
	if statePlainHash != localPlainHash {
		if err := d.SetNew("plain_content_hash", localPlainHash); err != nil {
			return err
		}
	}

	return nil
}

// loadFileContent returns contents of a file in a given path
func loadFileContent(v string) ([]byte, error) {
	filename, err := homedir.Expand(v)
	if err != nil {
		fmt.Printf("File %s can't be expand. %s", v, err)
		return nil, err
	}
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("File %s can't be read. %s", filename, err)
		return nil, err
	}
	return fileContent, nil
}

func getHash(data string) string {
	sha := sha256.New()
	sha.Write([]byte(data))
	shaSum := sha.Sum(nil)
	encoded := base64.StdEncoding.EncodeToString(shaSum[:])
	//h := md5.New()
	//io.WriteString(h, data)
	return encoded //fmt.Sprintf("%x", h.Sum(nil))
}
