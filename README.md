# calcdate

calcdate a an utility to make some basic operation on date.

```
Usage of calcdate:
  -b string
    	Begin date (default "// ::")
  -e string
    	End date
  -ifmt string
    	Input Format (YYYY/MM/DD hh:mm:ss) (default "YYYY/MM/DD hh:mm:ss")
  -ofmt string
    	Input Format (YYYY/MM/DD hh:mm:ss) (default "YYYY/MM/DD hh:mm:ss")
  -s string
    	Separator (default " ")
  -tz string
    	Timezone (default "Local")
```


By default this is the current datetime that is printed.

```
$ ./calcdate 
2020/09/14 23:15:10
```


# Some examples 

```
$ ./calcdate -b //-1      
2020/09/13 23:16:37
```

```
$ date && ./calcdate -b :-1: -e :-1:
lun. 14 sept. 2020 23:18:15 CEST
2020/09/14 23:17:00 2020/09/14 23:17:59
```

```
$ date && ./calcdate -b :-1:        
lun. 14 sept. 2020 23:18:22 CEST
2020/09/14 23:17:22
```

```
$ date && ./calcdate -b :-1: -tz UTC 
lun. 14 sept. 2020 23:19:08 CEST
2020/09/14 21:18:08
```


# Make a release

```
git tag -a v0.1.0 -m "First release"
git push origin v0.1.0
goreleaser --snapshot  # Check
goreleaser 
```
