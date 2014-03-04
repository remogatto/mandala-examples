// +build android

#include <android/native_activity.h>

void showAdPopup(ANativeActivity *activity)
{
  JavaVM* jvm = activity->vm;
  JNIEnv* env = NULL;
  (*jvm)->GetEnv(jvm, (void **)&env, JNI_VERSION_1_6);
  jint res = (*jvm)->AttachCurrentThread(jvm, &env, NULL);

  if (res == JNI_ERR) {
    // Failed to retrieve JVM environment
    return; 
  }

  jclass clazz = (*env)->GetObjectClass(env, activity->clazz);
  jmethodID methodID = (*env)->GetMethodID(env, clazz, "showAdPopup", "()V");
  (*env)->CallVoidMethod(env, activity->clazz, methodID);
  (*jvm)->DetachCurrentThread(jvm);
}
