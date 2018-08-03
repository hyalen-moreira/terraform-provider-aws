package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSKinesisAnalyticsApplication_basic(t *testing.T) {
	var application kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplication_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &application),
					resource.TestCheckResourceAttr(resName, "version", "1"),
					resource.TestCheckResourceAttr(resName, "code", "testCode\n"),
				),
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_update(t *testing.T) {
	var application kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplication_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &application),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplication_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "code", "testCode2\n"),
					resource.TestCheckResourceAttr(resName, "version", "3"),
					resource.TestCheckResourceAttr(resName, "cloudwatch_logging_options.#", "1"),
				),
			},
		},
	})
}

//func TestAccAWSKinesisAnalyticsApplication_cloudwatchLoggingOptions(t *testing.T) {
//	var application kinesisanalytics.ApplicationDetail
//	resName := "aws_kinesis_analytics_application.test"
//	rInt := acctest.RandInt()
//
//	resource.Test(t, resource.TestCase{
//		PreCheck:     func() { testAccPreCheck(t) },
//		Providers:    testAccProviders,
//		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccKinesisAnalyticsApplication_cloudwatchLoggingOptions(rInt),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckKinesisAnalyticsApplicationExists(resName, &application),
//					resource.TestCheckResourceAttr(resName, "cloudwatch_logging_options.#", "1"),
//				),
//			},
//		},
//	})
//}

func testAccCheckKinesisAnalyticsApplicationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesis_analytics_application" {
			continue
		}
		conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsconn
		describeOpts := &kinesisanalytics.DescribeApplicationInput{
			ApplicationName: aws.String(rs.Primary.Attributes["name"]),
		}
		resp, err := conn.DescribeApplication(describeOpts)
		if err == nil {
			if resp.ApplicationDetail != nil && *resp.ApplicationDetail.ApplicationStatus != kinesisanalytics.ApplicationStatusDeleting {
				return fmt.Errorf("Error: Application still exists")
			}
		}
		return nil
	}
	return nil
}

func testAccCheckKinesisAnalyticsApplicationExists(n string, application *kinesisanalytics.ApplicationDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis Analytics Application ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsconn
		describeOpts := &kinesisanalytics.DescribeApplicationInput{
			ApplicationName: aws.String(rs.Primary.Attributes["name"]),
		}
		resp, err := conn.DescribeApplication(describeOpts)
		if err != nil {
			return err
		}

		*application = *resp.ApplicationDetail

		return nil
	}
}

func testAccKinesisAnalyticsApplication_basic(rInt int) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type = "Service"
      identifiers = ["kinesisanalytics.amazonaws.com"]
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = "testAcc-%d"
}

resource "aws_cloudwatch_log_stream" "test" {
  name = "testAcc-%d"
  log_group_name = "${aws_cloudwatch_log_group.test.name}"
}

resource "aws_iam_role" "test" {
  name = "testAcc-%d"
  assume_role_policy = "${data.aws_iam_policy_document.test.json}" 
}

resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"
  code = "testCode\n"
}
`, rInt, rInt, rInt, rInt)
}

func testAccKinesisAnalyticsApplication_update(rInt int) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type = "Service"
      identifiers = ["kinesisanalytics.amazonaws.com"]
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = "testAcc-%d"
}

resource "aws_cloudwatch_log_stream" "test" {
  name = "testAcc-%d"
  log_group_name = "${aws_cloudwatch_log_group.test.name}"
}

resource "aws_iam_role" "test" {
  name = "testAcc-%d"
  assume_role_policy = "${data.aws_iam_policy_document.test.json}" 
}

resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"
  code = "testCode2\n"
  cloudwatch_logging_options {
    log_stream = "${aws_cloudwatch_log_stream.test.arn}"
    role = "${aws_iam_role.test.arn}"
  }
}`, rInt, rInt, rInt, rInt)
}
