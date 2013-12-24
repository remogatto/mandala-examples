#!/bin/bash

APKNAME=`ls src/`

cd android
ln -sT libs lib
zip -u bin/$APKNAME-debug.apk lib/armeabi/lib$APKNAME.so
rm lib
cd ..
./signapk.sh "android/bin/$APKNAME-debug.apk"
mv signed_$APKNAME-debug.apk "android/bin/$APKNAME-debug.apk"
