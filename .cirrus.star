load("github.com/SonarSource/cirrus-modules@v3", "load_features")
load(
    "github.com/SonarSource/cirrus-modules/cloud-native/helper.star@analysis/master",
    "merge_dict"
)

load(
    "github.com/SonarSource/cirrus-modules/cloud-native/conditions.star@analysis/master",
    "is_main_branch"
)

load(
    "github.com/SonarSource/cirrus-modules/cloud-native/env.star@analysis/master",
    "cirrus_env",
    "whitesource_api_env"
)
load(
    "github.com/SonarSource/cirrus-modules/cloud-native/platform.star@analysis/master",
    "custom_image_container_builder"
)


def main(ctx):
    conf = {}
    merge_dict(conf, load_features(ctx))
    merge_dict(conf, build_task())
    merge_dict(conf, sca_scan_task())
    return conf


def build_task():
    return {
        "build_task": {
            "env": {
                "CIRRUS_CLONE_DEPTH": 10,
                "GO_VERSION": "1.21.8",
            },
            "eks_container": custom_image_container_builder(cpu=1, memory="1G"),
            "modules_cache": {
                "fingerprint_script": "cat src/go.sum",
                "folder": "/home/sonarsource/go/pkg/mod"
            },
            "build_script": [
                "cd src",
                "go build -v ./...",
                "go test -v ./..."
            ]
        }
    }

#
# WhiteSource scan
#

def whitesource_script():
  return [
    "source cirrus-env QA",
    "export PROJECT_VERSION=$(cat VERSION | grep 'go[\\d.]*' | sed 's/go//').${BUILD_NUMBER:-0}",
    "source ws_scan.sh"
  ]


def sca_scan_task():
  return {
    "sca_scan_task": {
      # "only_if": is_main_branch(),
      "depends_on": "build",
      "env": whitesource_env(),
      "eks_container": custom_image_container_builder(cpu=1, memory="1G"),
      "whitesource_script": whitesource_script(),
      "allow_failures": "true",
      "always": {
        "ws_artifacts": {
          "path": "whitesource/**/*"
        }
      },
    }
  }

def whitesource_env():
  env = {
            "CIRRUS_CLONE_DEPTH": 10,
            "GO_VERSION": "1.21.8",
  }
  return whitesource_api_env() | env

