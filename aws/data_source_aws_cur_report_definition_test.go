package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsCurReportDefinition_basic(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	datasourceName := "data.aws_cur_report_definition.test"

	reportName := acctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	bucketRegion := "us-east-1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsCurReportDefinitionConfig_basic(reportName, bucketName, bucketRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsCurReportDefinitionCheckExists(datasourceName, resourceName),
					resource.TestCheckResourceAttr(datasourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(datasourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(datasourceName, "compression", "GZIP"),
					resource.TestCheckResourceAttr(datasourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(datasourceName, "s3_prefix", ""),
					resource.TestCheckResourceAttr(datasourceName, "s3_region", bucketRegion),
					resource.TestCheckResourceAttr(datasourceName, "additional_artifacts.#", "2"),
				),
			},
		},
	})
}

func testAccDataSourceAwsCurReportDefinitionCheckExists(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no data source called %s", datasourceName)
		}
		_, ok = s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}
		return nil
	}
}

// note: cur report definitions are currently only supported in us-east-1
func testAccDataSourceAwsCurReportDefinitionConfig_basic(reportName string, bucketName string, bucketRegion string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_s3_bucket" "test" {
	bucket = "%[2]s"
	acl = "private"
	force_destroy = true
    region = "%[3]s"
}

resource "aws_s3_bucket_policy" "test" {
  bucket = "${aws_s3_bucket.test.id}"
  policy = <<POLICY
{
    "Version": "2008-10-17",
    "Id": "s3policy",
    "Statement": [
        {
            "Sid": "AllowCURBillingACLPolicy",
            "Effect": "Allow",
            "Principal": {
                "AWS": "${data.aws_billing_service_account.test.arn}"
            },
            "Action": [
                "s3:GetBucketAcl",
                "s3:GetBucketPolicy"
            ],
            "Resource": "arn:aws:s3:::${aws_s3_bucket.test.id}"
        },
        {
            "Sid": "AllowCURPutObject",
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::386209384616:root"
            },
            "Action": "s3:PutObject",
            "Resource": "arn:aws:s3:::${aws_s3_bucket.test.id}/*"
        }
    ]
}
POLICY
}

resource "aws_cur_report_definition" "test" {
    report_name = "%[1]s"
    time_unit = "DAILY"
    format = "textORcsv"
    compression = "GZIP"
    additional_schema_elements = ["RESOURCES"]
    s3_bucket = "${aws_s3_bucket.test.id}"
    s3_prefix = ""
    s3_region = "${aws_s3_bucket.test.region}"
	additional_artifacts = ["REDSHIFT", "QUICKSIGHT"]
}

data "aws_cur_report_definition" "test" {
    report_name = "${aws_cur_report_definition.test.report_name}"
}
`, reportName, bucketName, bucketRegion)
}
