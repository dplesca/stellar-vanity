stellar vanity address generator
====

This project is simple stellar vanity address generator written in go. Its usage is as easy as:
```
./stellar-vanity text
```
It has however a few neat tricks up its sleeve: 
 - it uses goroutines to concurrently generate as many keypairs as it can (but not too many, as it's just wasteful). You can of course raise or lower the `maxConcurrency` value on build. 
 - it can be used to search for a text: anywhere (default), start or end of the address.
 - it can show you some progress information, it prints out a message every 100,000 generated keypairs (using `--verbose` flag)
 - it will show you where the text has been found in the address in a colorful way
 - it can output the result in a text file too, just be careful with it, please (using `--writetofile` flag)

A small note for searching a string on the start of the address. The first two letters of a stellar address can be: GA, GB, GC, GD. The reason is [a side effect of base32 encoding according to this stackexchange answer](https://stellar.stackexchange.com/questions/371/does-the-second-letter-of-the-public-address-having-any-meaning-since-it-only-ap). (because the of the beta status, I'll paste the relevant part below)

> the first byte (8 bits) that is encoded contains the type of the string. A public key has prefix "G" for example.  
> when converting into base32 the data is consumed 5 bits at a time, so the first 5 bits of the 8 bits version end up being the first character. The second character is therefore the remaining 3 bits from the version byte (but they are all 0s), plus the first 2 bits of the actual data. 2 bits of data give you the characters A through D.

This vanity address finder will search for the requested string _as soon as it can_, meaning that if your text starts with *GA* it will look for a match right at the start of the address, but if your text starts with *B* it will look for a match starting from the second character of the address and, finally, if your text starts with *Q* it will look for a match starting from the third.

