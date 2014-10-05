# Proof-Of-Concept Demo Black-Box Test

The demo is tested using a black-box approach. The test passes when
the actual image produced by the demo is similar to an expected
image within a tolerance interval. The package
[imagetest](https://github.com/remogatto/imagetest) is used for image
comparaison.

# Usage

To run the tests on xorg and/or on Android you need a working
[Mandala](https://github.com/remogatto/mandala) environment.

<pre>
gotask test xorg # or
gotask test android
</pre>

# LICENSE

See [LICENSE](LICENSE).
