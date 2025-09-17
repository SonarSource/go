load("github.com/SonarSource/cirrus-modules@v3", "load_features")
load(
    "github.com/SonarSource/cirrus-modules/cloud-native/helper.star@analysis/master",
    "merge_dict"
)

load(
    "github.com/SonarSource/cirrus-modules/cloud-native/conditions.star@analysis/master",
    "is_main_branch",
    "is_branch_qa_eligible"
)

load(
    "github.com/SonarSource/cirrus-modules/cloud-native/env.star@analysis/master",
    "cirrus_env"
)
load(
    "github.com/SonarSource/cirrus-modules/cloud-native/platform.star@analysis/master",
    "custom_image_container_builder"
)


def main(ctx):
    conf = {}
    merge_dict(conf, load_features(ctx))
    merge_dict(conf, build_task())
    merge_dict(conf, shadow_scan_sqc_eu_task())
    merge_dict(conf, shadow_scan_sqc_us_task())
    return conf

def build_env():
    env = common_env()
    env |= {
       "SONAR_TOKEN": "VAULT[development/kv/data/next data.token]",
       "SONAR_HOST_URL": "https://next.sonarqube.com/sonarqube"
    }
    return env

def build_task():
    return {
        "build_task": {
            "env": build_env(),
            "eks_container": custom_image_container_builder(cpu=1, memory="1G"),
            "modules_cache": {
                "fingerprint_script": "cat src/go.sum",
                "folder": "/home/sonarsource/go/pkg/mod"
            },
            "build_script": [
                "cd src",
                "go build -v ./...",
                "go test -v ./... -coverprofile=coverage.out -json > test-report.out",
                "../.cirrus/analyze.sh"
            ]
        }
    }

#
# Shadow Scans
#

def is_run_shadow_scan():
    return "($CIRRUS_CRON == $CRON_NIGHTLY_JOB_NAME && $CIRRUS_BRANCH == \"master\") || $CIRRUS_PR_LABELS =~ \".*shadow_scan.*\""


def shadow_scan_task_template(env):
    return {
        "only_if": "({}) && ({})".format(is_branch_qa_eligible(), is_run_shadow_scan()),
        "depends_on": "build",
        "env": env,
        "eks_container": custom_image_container_builder(cpu=1, memory="1G"),
        "modules_cache": {
            "fingerprint_script": "cat src/go.sum",
            "folder": "/home/sonarsource/go/pkg/mod"
        },
        "build_script": [
            "cd src",
            "go build -v ./...",
            "go test -v ./... -coverprofile=coverage.out -json > test-report.out",
            "../.cirrus/analyze.sh"
        ]
    }

def shadow_scan_sqc_eu_env():
    env = common_env()
    env |= {
       "SONAR_TOKEN": "VAULT[development/kv/data/sonarcloud data.token]",
       "SONAR_HOST_URL": "https://sonarcloud.io"
    }
    return env

def shadow_scan_sqc_eu_task():
    return {
        "shadow_scan_sqc_eu_task": shadow_scan_task_template(shadow_scan_sqc_eu_env())
    }

def shadow_scan_sqc_us_env():
    env = common_env()
    env |= {
       "SONAR_TOKEN": "VAULT[development/kv/data/sonarqube-us data.token]",
       "SONAR_HOST_URL": "https://sonarqube.us"
    }
    return env

def shadow_scan_sqc_us_task():
    return {
        "shadow_scan_sqc_us_task": shadow_scan_task_template(shadow_scan_sqc_us_env())
    }


def common_env():
  return {
    "CIRRUS_CLONE_DEPTH": 10,
    "GO_VERSION": "1.25.1",
    "CRON_NIGHTLY_JOB_NAME": "nightly",
  }
