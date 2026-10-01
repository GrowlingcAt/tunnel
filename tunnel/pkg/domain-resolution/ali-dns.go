package domain_resolution

import (
	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"tunnel/pkg/log"
)

type aliDns struct {
	client     *alidns20150109.Client
	rootDomain string
}

func (f *dnsFactory) NewDns() IDomainResolution {
	client, err := createClient(f.config.AliYunDomain.AccessKeyID, f.config.AliYunDomain.AccessKeySecret, f.config.AliYunDomain.Endpoint)
	if err != nil {
		log.Error(err)
		return nil
	}
	return &aliDns{
		client:     client,
		rootDomain: f.config.AliYunDomain.RootDomain,
	}
}

func createClient(accessKeyID, accessKeySecret, endpoint string) (*alidns20150109.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(accessKeyID),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	}
	client, err := alidns20150109.NewClient(config)
	return client, err
}

func (d *aliDns) CheckSubDomainExists(subDomain string) (bool, error) {
	describeSubDomainRecordsRequest := &alidns20150109.DescribeSubDomainRecordsRequest{
		SubDomain:  tea.String(subDomain),
		DomainName: tea.String(d.rootDomain),
	}
	runtime := &util.RuntimeOptions{}
	res, err := d.client.DescribeSubDomainRecordsWithOptions(describeSubDomainRecordsRequest, runtime)
	if err != nil {
		log.Error(err)
		return false, err
	}
	if *res.Body.TotalCount > 0 {
		return true, nil
	}
	return false, nil
}
func (d *aliDns) AddSubDomainRecord(rr, typ, value string) (string, error) {
	addDomainRecordRequest := &alidns20150109.AddDomainRecordRequest{
		DomainName: tea.String(d.rootDomain),
		RR:         tea.String(rr),
		Type:       tea.String(typ),
		Value:      tea.String(value),
	}
	runtime := &util.RuntimeOptions{}
	res, err := d.client.AddDomainRecordWithOptions(addDomainRecordRequest, runtime)
	if err != nil {
		log.Error(err)
		return "", err
	}
	return *res.Body.RecordId, nil
}
func (d *aliDns) UpdateSubDomainRemark(recordID, remark string) error {
	updateDomainRecordRemarkRequest := &alidns20150109.UpdateDomainRecordRemarkRequest{
		RecordId: tea.String(recordID),
		Remark:   tea.String(remark),
	}
	runtime := &util.RuntimeOptions{}
	_, err := d.client.UpdateDomainRecordRemarkWithOptions(updateDomainRecordRemarkRequest, runtime)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
