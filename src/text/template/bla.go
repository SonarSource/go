import (
	"crypto"
	"crypto/md5"
	"crypto/sha1"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"
)

func main() {
	tmp := md4.New() // Noncompliant {{Make sure this weak hash algorithm is not used in a sensitive context here.}}
	//     ^^^^^^^^^
	ripemd160.New()                   // Noncompliant
	crypto.Hash.New(crypto.MD4)       // Noncompliant
	crypto.Hash.New(crypto.MD5)       // Noncompliant
	crypto.Hash.New(crypto.SHA1)      // Noncompliant
	crypto.Hash.New(crypto.RIPEMD160) // Noncompliant
	md5.Sum(data)                     // Noncompliant
	sha1.Sum(data)                    // Noncompliant

	md4.New(someArgs)                     // Noncompliant
	md5.Sum()                             // Noncompliant
	md5.Sum(more, args)                   // Noncompliant
	crypto.Hash.New(crypto.MD4, extraArg) // Noncompliant

	md4.other()
	Hash.New(crypto.MD4)
	crypto.Hash.New(crypto.SHA256)
	crypto.Hash.New(unknownValue)
	crypto.Hash.New()

	// Extra test case for code coverage of MethodCall
	"test".string()
	"test".string
	test.string
}
