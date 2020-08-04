#!/bin/bash
#
# Copyright 2020 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Get the SHA from the RepoDigests section of an operand image.
# Do "docker login" before running this script.
# Run this script from the parent dir by typing "scripts/get-image-sha.sh"

FILE=deploy/operator.yaml

# check the input parms
REGISTRY=$1
NAME=$2
TAG=$3
TYPE=$4
if [[ $TYPE == "" ]]
then
   echo "Missing parm. Need image registry, image name, image tag, and env variable indicating operand type. Tye will be OPERATOR for operator image as input"
   echo "for eg: quay.io/opencloudio icp-cert-manager-controller x.y.z CONTROLLER_IMAGE_TAG_OR_SHA"
   exit 1
fi

# pull the image
IMAGE="$REGISTRY/$NAME:$TAG"
echo "Pulling image $IMAGE"
docker pull "$IMAGE" &>/dev/null

# get the SHA for the image
DIGEST="$(docker images --digests "$REGISTRY"/"$NAME" |grep "$TAG" |awk 'FNR==1{print $3}')"

# DIGEST should look like this: eg: sha256:10a844ffaf7733176e927e6c4faa04c2bc4410cf4d4ef61b9ae5240aa62d1456
if [[ $DIGEST != sha256* ]]
then
    echo "Cannot find SHA (sha256:<DIGEST_SOME_HEX_VALUE>) in digest: $DIGEST"
    exit 1
fi

SHA=$DIGEST

echo "$NAME : $SHA"

# delete the "name" and "value" lines for the old SHA ONLY FOR OPERANDS
# for example:
#     - name: CONTROLLER_IMAGE_TAG_OR_SHA
#       value: "sha256:10a844ffaf7733176e927e6c4faa04c2bc4410cf4d4ef61b9ae5240aa62d1456"

sed -i "/name: $TYPE/{N;d;}" $FILE

# insert the new SHA lines
LINE1="\            - name: $TYPE"
LINE2="\              value: \"$SHA\""
sed -i "/DO NOT DELETE. Add operand image SHAs here./a $LINE1\n$LINE2" $FILE
