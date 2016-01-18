#!/bin/bash

# change to root of bosh release
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $DIR/../..

cat > ~/.bosh_config << EOF
---
aliases:
  target:
    bosh-lite: ${bosh_target}
auth:
  ${bosh_target}:
    username: ${bosh_username}
    password: ${bosh_password}
EOF
bosh target ${bosh_target}

_bosh() {
  bosh -n $@
}

set -e
git submodule update --init --recursive --force

_bosh create release --force
set +e
_bosh upload release --rebase || echo "Continuing..."
set -e

cd ../
git clone https://github.com/cloudfoundry-community/postgres-smoke-tests-boshrelease.git 
cd postgres-smoke-tests-boshrelease
set -e 
git submodule update --init --recursive --force
_bosh create release --force
set +e
_bosh upload release --rebase || echo "Continuing..."
set -e

cd $DIR/../../ 

echo "Creating stub"

cat > ci/aws/templates/${environment}/stub.yml << EOF
---
meta:
  environment: ${bosh_deployment_name}
  stemcell:
    name: bosh-aws-xen-centos-7-go_agent
    version: latest

director_uuid: ${bosh_uuid}

releases:
  - name: rdpg
    version: latest

properties: {}
EOF

spiff merge templates/deployment.yml \
            templates/jobs.yml \
            ci/aws/templates/smoke_tests.yml \
            templates/infrastructure/aws.yml \
            ci/aws/templates/${environment}/networks.yml \
            ci/aws/templates/${environment}/properties.yml \
            ci/aws/templates/overwrites.yml \
            ci/aws/templates/${environment}/stub.yml > manifests/aws.yml

_bosh deployment manifests/aws.yml
_bosh deploy
