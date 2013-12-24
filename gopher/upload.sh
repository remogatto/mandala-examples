#!/bin/bash

set -e

./build.sh
./refresh.sh
adb install -r android/bin/cube-debug.apk
