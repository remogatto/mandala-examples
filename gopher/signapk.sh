#!/bin/bash
#
# Original signapk.sh script from https://code.google.com/p/apk-resigner/
#
# Sample usage is as follows;
# ./signapk myapp.apk debug.keystore android androiddebugkey
# 
# param1, APK file: Calculator_debug.apk
# param2, keystore location: ~/.android/debug.keystore
# param3, key storepass: android
# param4, key alias: androiddebugkey

USER_HOME=$(eval echo ~${SUDO_USER})

# use my debug key default
APK=$1
KEYSTORE="${2:-$USER_HOME/.android/debug.keystore}"
STOREPASS="${3:-android}"
ALIAS="${4:-androiddebugkey}"

APK_BASENAME=$(basename $APK)
SIGNED_APK="signed_"$APK_BASENAME

# delete META-INF folder
zip -d $APK META-INF/\*

# sign APK
jarsigner -sigalg SHA1withRSA -digestalg SHA1 -keystore $KEYSTORE -storepass $STOREPASS $APK $ALIAS

#verify
#jarsigner -verify $APK

#zipalign
zipalign 4 $APK $SIGNED_APK
