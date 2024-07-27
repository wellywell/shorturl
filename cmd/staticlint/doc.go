package main

/*staticlint is a tool for static analysis of Go programs.

staticlint examines Go source code and reports suspicious constructs,
such as Printf calls whose arguments do not align with the format
string. It uses heuristics that do not guarantee all reports are
genuine problems, but it can find errors not caught by the compilers.

Registered analyzers:

    SA1000       Invalid regular expression
    SA1001       Invalid template
    SA1002       Invalid format in time.Parse
    SA1003       Unsupported argument to functions in encoding/binary
    SA1004       Suspiciously small untyped constant in time.Sleep
    SA1005       Invalid first argument to exec.Command
    SA1006       Printf with dynamic first argument and no further arguments
    SA1007       Invalid URL in net/url.Parse
    SA1008       Non-canonical key in http.Header map
    SA1010       (*regexp.Regexp).FindAll called with n == 0, which will always return zero results
    SA1011       Various methods in the 'strings' package expect valid UTF-8, but invalid input is provided
    SA1012       A nil context.Context is being passed to a function, consider using context.TODO instead
    SA1013       io.Seeker.Seek is being called with the whence constant as the first argument, but it should be the second
    SA1014       Non-pointer value passed to Unmarshal or Decode
    SA1015       Using time.Tick in a way that will leak. Consider using time.NewTicker, and only use time.Tick in tests, commands and endless functions
    SA1016       Trapping a signal that cannot be trapped
    SA1017       Channels used with os/signal.Notify should be buffered
    SA1018       strings.Replace called with n == 0, which does nothing
    SA1019       Using a deprecated function, variable, constant or field
    SA1020       Using an invalid host:port pair with a net.Listen-related function
    SA1021       Using bytes.Equal to compare two net.IP
    SA1023       Modifying the buffer in an io.Writer implementation
    SA1024       A string cutset contains duplicate characters
    SA1025       It is not possible to use (*time.Timer).Reset's return value correctly
    SA1026       Cannot marshal channels or functions
    SA1027       Atomic access to 64-bit variable must be 64-bit aligned
    SA1028       sort.Slice can only be used on slices
    SA1029       Inappropriate key in call to context.WithValue
    SA1030       Invalid argument in call to a strconv function
    SA2000       sync.WaitGroup.Add called inside the goroutine, leading to a race condition
    SA2001       Empty critical section, did you mean to defer the unlock?
    SA2002       Called testing.T.FailNow or SkipNow in a goroutine, which isn't allowed
    SA2003       Deferred Lock right after locking, likely meant to defer Unlock instead
    SA3000       TestMain doesn't call os.Exit, hiding test failures
    SA3001       Assigning to b.N in benchmarks distorts the results
    SA4000       Binary operator has identical expressions on both sides
    SA4001       &*x gets simplified to x, it does not copy x
    SA4003       Comparing unsigned values against negative values is pointless
    SA4004       The loop exits unconditionally after one iteration
    SA4005       Field assignment that will never be observed. Did you mean to use a pointer receiver?
    SA4006       A value assigned to a variable is never read before being overwritten. Forgotten error check or dead code?
    SA4008       The variable in the loop condition never changes, are you incrementing the wrong variable?
    SA4009       A function argument is overwritten before its first use
    SA4010       The result of append will never be observed anywhere
    SA4011       Break statement with no effect. Did you mean to break out of an outer loop?
    SA4012       Comparing a value against NaN even though no value is equal to NaN
    SA4013       Negating a boolean twice (!!b) is the same as writing b. This is either redundant, or a typo.
    SA4014       An if/else if chain has repeated conditions and no side-effects; if the condition didn't match the first time, it won't match the second time, either
    SA4015       Calling functions like math.Ceil on floats converted from integers doesn't do anything useful
    SA4016       Certain bitwise operations, such as x ^ 0, do not do anything useful
    SA4017       Discarding the return values of a function without side effects, making the call pointless
    SA4018       Self-assignment of variables
    SA4019       Multiple, identical build constraints in the same file
    SA4020       Unreachable case clause in a type switch
    SA4021       'x = append(y)' is equivalent to 'x = y'
    SA4022       Comparing the address of a variable against nil
    SA4023       Impossible comparison of interface value with untyped nil
    SA4024       Checking for impossible return value from a builtin function
    SA4025       Integer division of literals that results in zero
    SA4026       Go constants cannot express negative zero
    SA4027       (*net/url.URL).Query returns a copy, modifying it doesn't change the URL
    SA4028       x % 1 is always zero
    SA4029       Ineffective attempt at sorting slice
    SA4030       Ineffective attempt at generating random number
    SA4031       Checking never-nil value against nil
    SA5000       Assignment to nil map
    SA5001       Deferring Close before checking for a possible error
    SA5002       The empty for loop ('for {}') spins and can block the scheduler
    SA5003       Defers in infinite loops will never execute
    SA5004       'for { select { ...' with an empty default branch spins
    SA5005       The finalizer references the finalized object, preventing garbage collection
    SA5007       Infinite recursive call
    SA5008       Invalid struct tag
    SA5009       Invalid Printf call
    SA5010       Impossible type assertion
    SA5011       Possible nil pointer dereference
    SA5012       Passing odd-sized slice to function expecting even size
    SA6000       Using regexp.Match or related in a loop, should use regexp.Compile
    SA6001       Missing an optimization opportunity when indexing maps by byte slices
    SA6002       Storing non-pointer values in sync.Pool allocates memory
    SA6003       Converting a string to a slice of runes before ranging over it
    SA6005       Inefficient string comparison with strings.ToLower or strings.ToUpper
    SA9001       Defers in range loops may not run when you expect them to
    SA9002       Using a non-octal os.FileMode that looks like it was meant to be in octal.
    SA9003       Empty body in an if or else branch
    SA9004       Only the first constant has an explicit type
    SA9005       Trying to marshal a struct with no public fields nor custom marshaling
    SA9006       Dubious bit shifting of a fixed size integer value
    SA9007       Deleting a directory that shouldn't be deleted
    SA9008       else branch of a type assertion is probably not reading the right value
    ST1000       Incorrect or missing package comment
    ST1001       Dot imports are discouraged
    ST1003       Poorly chosen identifier
    ST1005       Incorrectly formatted error string
    ST1006       Poorly chosen receiver name
    ST1008       A function's error value should be its last return value
    ST1011       Poorly chosen name for variable of type time.Duration
    ST1012       Poorly chosen name for error variable
    ST1013       Should use constants for HTTP error codes, not magic numbers
    ST1015       A switch's default case should be the first or last case
    ST1016       Use consistent method receiver names
    ST1017       Don't use Yoda conditions
    ST1018       Avoid zero-width and control characters in string literals
    ST1019       Importing the same package multiple times
    ST1020       The documentation of an exported function should start with the function's name
    ST1021       The documentation of an exported type should start with type's name
    ST1022       The documentation of an exported variable or constant should start with variable's name
    ST1023       Redundant type in variable declaration
    appends      check for missing values after append
    assign       check for useless assignments
    atomic       check for common mistakes using the sync/atomic package
    atomicalign  check for non-64-bits-aligned arguments to sync/atomic functions
    bools        check for common mistakes involving boolean operators
    composites   check for unkeyed composite literals
    copylocks    check for locks erroneously passed by value
    deepequalerrors check for calls of reflect.DeepEqual on error values
    defers       report common mistakes in defer statements
    directive    check Go toolchain directives such as //go:debug
    errcheck     check for unchecked errors
    errorsas     report passing non-pointer or non-error values to errors.As
    fieldalignment find structs that would use less memory if their fields were sorted
    httpmux      report using Go 1.22 enhanced ServeMux patterns in older Go versions
    httpresponse check for mistakes using HTTP responses
    ifaceassert  detect impossible interface-to-interface type assertions
    loopclosure  check references to loop variables from within nested functions
    lostcancel   check cancel func returned by context.WithCancel is called
    nilfunc      check for useless comparisons between functions and nil
    noosexit     check for os.Exit() in main() function of main package
    printf       check consistency of Printf format strings and arguments
    reflectvaluecompare check for comparing reflect.Value values with == or reflect.DeepEqual
    shadow       check for possible unintended shadowing of variables
    shift        check for shifts that equal or exceed the width of the integer
    stdmethods   check signature of methods of well-known interfaces
    stdversion   report uses of too-new standard library symbols
    stringintconv check for string(int) conversions
    structtag    check that struct field tags conform to reflect.StructTag.Get
    testifylint  Checks usage of github.com/stretchr/testify.
    testinggoroutine report calls to (*testing.T).Fatal from goroutines started by a test
    tests        check for common mistaken usages of tests and examples
    timeformat   check for calls of (time.Time).Format or time.Parse with 2006-02-01
    unmarshal    report passing non-pointer or non-interface values to unmarshal
    unreachable  check for unreachable code
    unusedresult check for unused results of calls to some functions
    unusedwrite  checks for unused writes
    usesgenerics detect whether a package uses generics features

By default all analyzers are run.
To select specific analyzers, use the -NAME flag for each one,
 or -NAME=false to run all analyzers not explicitly disabled.

Core flags:

  -SA1000
        enable SA1000 analysis
  -SA1001
        enable SA1001 analysis
  -SA1002
        enable SA1002 analysis
  -SA1003
        enable SA1003 analysis
  -SA1004
        enable SA1004 analysis
  -SA1005
        enable SA1005 analysis
  -SA1006
        enable SA1006 analysis
  -SA1007
        enable SA1007 analysis
  -SA1008
        enable SA1008 analysis
  -SA1010
        enable SA1010 analysis
  -SA1011
        enable SA1011 analysis
  -SA1012
        enable SA1012 analysis
  -SA1013
        enable SA1013 analysis
  -SA1014
        enable SA1014 analysis
  -SA1015
        enable SA1015 analysis
  -SA1016
        enable SA1016 analysis
  -SA1017
        enable SA1017 analysis
  -SA1018
        enable SA1018 analysis
  -SA1019
        enable SA1019 analysis
  -SA1020
        enable SA1020 analysis
  -SA1021
        enable SA1021 analysis
  -SA1023
        enable SA1023 analysis
  -SA1024
        enable SA1024 analysis
  -SA1025
        enable SA1025 analysis
  -SA1026
        enable SA1026 analysis
  -SA1027
        enable SA1027 analysis
  -SA1028
        enable SA1028 analysis
  -SA1029
        enable SA1029 analysis
  -SA1030
        enable SA1030 analysis
  -SA2000
        enable SA2000 analysis
  -SA2001
        enable SA2001 analysis
  -SA2002
        enable SA2002 analysis
  -SA2003
        enable SA2003 analysis
  -SA3000
        enable SA3000 analysis
  -SA3001
        enable SA3001 analysis
  -SA4000
        enable SA4000 analysis
  -SA4001
        enable SA4001 analysis
  -SA4003
        enable SA4003 analysis
  -SA4004
        enable SA4004 analysis
  -SA4005
        enable SA4005 analysis
  -SA4006
        enable SA4006 analysis
  -SA4008
        enable SA4008 analysis
  -SA4009
        enable SA4009 analysis
  -SA4010
        enable SA4010 analysis
  -SA4011
        enable SA4011 analysis
  -SA4012
        enable SA4012 analysis
  -SA4013
        enable SA4013 analysis
  -SA4014
        enable SA4014 analysis
  -SA4015
        enable SA4015 analysis
  -SA4016
        enable SA4016 analysis
  -SA4017
        enable SA4017 analysis
  -SA4018
        enable SA4018 analysis
  -SA4019
        enable SA4019 analysis
  -SA4020
        enable SA4020 analysis
  -SA4021
        enable SA4021 analysis
  -SA4022
        enable SA4022 analysis
  -SA4023
        enable SA4023 analysis
  -SA4024
        enable SA4024 analysis
  -SA4025
        enable SA4025 analysis
  -SA4026
        enable SA4026 analysis
  -SA4027
        enable SA4027 analysis
  -SA4028
        enable SA4028 analysis
  -SA4029
        enable SA4029 analysis
  -SA4030
        enable SA4030 analysis
  -SA4031
        enable SA4031 analysis
  -SA5000
        enable SA5000 analysis
  -SA5001
        enable SA5001 analysis
  -SA5002
        enable SA5002 analysis
  -SA5003
        enable SA5003 analysis
  -SA5004
        enable SA5004 analysis
  -SA5005
        enable SA5005 analysis
  -SA5007
        enable SA5007 analysis
  -SA5008
        enable SA5008 analysis
  -SA5009
        enable SA5009 analysis
  -SA5010
        enable SA5010 analysis
  -SA5011
        enable SA5011 analysis
  -SA5012
        enable SA5012 analysis
  -SA6000
        enable SA6000 analysis
  -SA6001
        enable SA6001 analysis
  -SA6002
        enable SA6002 analysis
  -SA6003
        enable SA6003 analysis
  -SA6005
        enable SA6005 analysis
  -SA9001
        enable SA9001 analysis
  -SA9002
        enable SA9002 analysis
  -SA9003
        enable SA9003 analysis
  -SA9004
        enable SA9004 analysis
  -SA9005
        enable SA9005 analysis
  -SA9006
        enable SA9006 analysis
  -SA9007
        enable SA9007 analysis
  -SA9008
        enable SA9008 analysis
  -ST1000
        enable ST1000 analysis
  -ST1001
        enable ST1001 analysis
  -ST1003
        enable ST1003 analysis
  -ST1005
        enable ST1005 analysis
  -ST1006
        enable ST1006 analysis
  -ST1008
        enable ST1008 analysis
  -ST1011
        enable ST1011 analysis
  -ST1012
        enable ST1012 analysis
  -ST1013
        enable ST1013 analysis
  -ST1015
        enable ST1015 analysis
  -ST1016
        enable ST1016 analysis
  -ST1017
        enable ST1017 analysis
  -ST1018
        enable ST1018 analysis
  -ST1019
        enable ST1019 analysis
  -ST1020
        enable ST1020 analysis
  -ST1021
        enable ST1021 analysis
  -ST1022
        enable ST1022 analysis
  -ST1023
        enable ST1023 analysis
  -V    print version and exit
  -all
        no effect (deprecated)
  -appends
        enable appends analysis
  -assign
        enable assign analysis
  -atomic
        enable atomic analysis
  -atomicalign
        enable atomicalign analysis
  -bool
        deprecated alias for -bools
  -bools
        enable bools analysis
  -c int
        display offending line with this many lines of context (default -1)
  -composites
        enable composites analysis
  -compositewhitelist
        deprecated alias for -composites.whitelist (default true)
  -copylocks
        enable copylocks analysis
  -cpuprofile string
        write CPU profile to this file
  -debug string
        debug flags, any subset of "fpstv"
  -deepequalerrors
        enable deepequalerrors analysis
  -defers
        enable defers analysis
  -directive
        enable directive analysis
  -errcheck
        enable errcheck analysis
  -errorsas
        enable errorsas analysis
  -fieldalignment
        enable fieldalignment analysis
  -fix
        apply all suggested fixes
  -flags
        print analyzer flags in JSON
  -httpmux
        enable httpmux analysis
  -httpresponse
        enable httpresponse analysis
  -ifaceassert
        enable ifaceassert analysis
  -json
        emit JSON output
  -loopclosure
        enable loopclosure analysis
  -lostcancel
        enable lostcancel analysis
  -memprofile string
        write memory profile to this file
  -methods
        deprecated alias for -stdmethods
  -nilfunc
        enable nilfunc analysis
  -noosexit
        enable noosexit analysis
  -printf
        enable printf analysis
  -printfuncs value
        deprecated alias for -printf.funcs (default (*log.Logger).Fatal,(*log.Logger).Fatalf,(*log.Logger).Fatalln,(*log.Logger).Panic,(*log.Logger).Panicf,(*log.Logger).Panicln,(*log.Logger).Print,(*log.Logger).Printf,(*log.Logger).Println,(*testing.common).Error,(*testing.common).Errorf,(*testing.common).Fatal,(*testing.common).Fatalf,(*testing.common).Log,(*testing.common).Logf,(*testing.common).Skip,(*testing.common).Skipf,(testing.TB).Error,(testing.TB).Errorf,(testing.TB).Fatal,(testing.TB).Fatalf,(testing.TB).Log,(testing.TB).Logf,(testing.TB).Skip,(testing.TB).Skipf,fmt.Append,fmt.Appendf,fmt.Appendln,fmt.Errorf,fmt.Fprint,fmt.Fprintf,fmt.Fprintln,fmt.Print,fmt.Printf,fmt.Println,fmt.Sprint,fmt.Sprintf,fmt.Sprintln,log.Fatal,log.Fatalf,log.Fatalln,log.Panic,log.Panicf,log.Panicln,log.Print,log.Printf,log.Println,runtime/trace.Logf)
  -rangeloops
        deprecated alias for -loopclosure
  -reflectvaluecompare
        enable reflectvaluecompare analysis
  -shadow
        enable shadow analysis
  -shadowstrict
        deprecated alias for -shadow.strict
  -shift
        enable shift analysis
  -source
        no effect (deprecated)
  -stdmethods
        enable stdmethods analysis
  -stdversion
        enable stdversion analysis
  -stringintconv
        enable stringintconv analysis
  -structtag
        enable structtag analysis
  -tags string
        no effect (deprecated)
  -test
        indicates whether test files should be analyzed, too (default true)
  -testifylint
        enable testifylint analysis
  -testinggoroutine
        enable testinggoroutine analysis
  -tests
        enable tests analysis
  -timeformat
        enable timeformat analysis
  -trace string
        write trace log to this file
  -unmarshal
        enable unmarshal analysis
  -unreachable
        enable unreachable analysis
  -unusedfuncs value
        deprecated alias for -unusedresult.funcs (default context.WithCancel,context.WithDeadline,context.WithTimeout,context.WithValue,errors.New,fmt.Errorf,fmt.Sprint,fmt.Sprintf,slices.Clip,slices.Compact,slices.CompactFunc,slices.Delete,slices.DeleteFunc,slices.Grow,slices.Insert,slices.Replace,sort.Reverse)
  -unusedresult
        enable unusedresult analysis
  -unusedstringmethods value
        deprecated alias for -unusedresult.stringmethods (default Error,String)
  -unusedwrite
        enable unusedwrite analysis
  -usesgenerics
        enable usesgenerics analysis
  -v    no effect (deprecated)

To see details and flags of a specific analyzer, run 'staticlint help name'. */
