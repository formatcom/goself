~~~
$ mkdir /tmp/vinicio
$ sudo ./goself --help
$ sudo ./goself --unshare --fork
$ sudo umount /tmp/vinicio
~~~

~~~
[./goself --unshare --fork]
[]
INIT LOADER 46923
     SET CLONE_NEWNS

[init --unshare --fork loader]
[loader]
     LOADER 46927
        test 1 [/tmp/vinicio] -> ls 
        test 1 [] -> vicky

     FORK EXEC

[init unshare]
[unshare]
     UNSHARE 46931
        mkdir /tmp/vinicio/test
        err mkdir ->  <nil>
        test 2 [/tmp/vinicio] -> ls 
        test 2 [] -> test
        test 2 [] -> vicky

        mount /tmp/vinicio
        err mount ->  <nil>
        err mount ->  <nil>
        test 3 [/tmp/vinicio] -> ls 

        test 4 [/tmp/vinicio] -> ls goroutine

        mkdir /tmp/vinicio/test1
        err mkdir ->  <nil>
        mkdir /tmp/vinicio/test2
        err mkdir ->  <nil>
        mkdir /tmp/vinicio/test3
        err mkdir ->  <nil>

        test 5 [/tmp/vinicio] -> ls with data
        test 5 [with data] -> test1
        test 5 [with data] -> test2
        test 5 [with data] -> test3

        test 6 [/tmp/vinicio] -> ls with data goroutine
        test 6 [with data goroutine] -> test1
        test 6 [with data goroutine] -> test2
        test 6 [with data goroutine] -> test3

        test 7 [/tmp/vinicio] -> ls with data goroutine
        test 7 [with data goroutine] -> test1
        test 7 [with data goroutine] -> test2
        test 7 [with data goroutine] -> test3

     FAKE REQUEST 46931

     ENGINE 46938

  mkdir /tmp/vinicio/test
  err mkdir ->  <nil>
        test 8 [/tmp/vinicio] -> ls 
        test 8 [] -> test
        test 8 [] -> test1
        test 8 [] -> test2
        test 8 [] -> test3

        test 9 [/tmp/vinicio/test] -> ls 

  mount /tmp/vinicio/test
  err mount ->  <nil>
  err mount ->  <nil>
  mkdir /tmp/vinicio/test/vinicio
  err mkdir ->  <nil>
        test 10 [/tmp/vinicio/test] -> ls 
        test 10 [] -> in.out
        test 10 [] -> vinicio

python -> ls /tmp/vinicio
 - test
 - test3
 - test2
 - test1
python -> ls /tmp/vinicio/test
 - in.out
 - vinicio
python -> Goodbye, World!
     END ENGINE

        test 11 [/tmp/vinicio/test] -> ls 

        test 12 [/tmp/vinicio/test] -> ls 

        test 13 [/tmp/vinicio/test] -> ls 
~~~
