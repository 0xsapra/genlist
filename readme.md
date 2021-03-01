
# Usage 

* genlist -w dirlist1.txt -w dirlist2.txt -e php,txt,bin -t tranform1.txt -t transform2.txt -d https://www.site.com -ssrf http://zeta2.beeceptor.com/probe -of disearch
* genlist -w dirlist1.txt -w dirlist2.txt -e php,txt,bin -t tranform1.txt -t transform2.txt -d https://site.com -ssrf http://zeta2.beeceptor.com/probe -of ffuf
* genlist -w dirlist1.txt -w dirlist2.txt -e php,txt,bin -t tranform1.txt -t transform2.txt -d https://sapra.site.com -ssrf http://zeta2.beeceptor.com/probe


# Flags

| Flag  | Description  | Example |
| ----------- | ----------- | ----------- |
| -of |  output format(can be ffuf or dirsearch) | -of dirsearch |
| -w |  input wordlists | -w inputwords.txt |
| -e |  extensions | -e php,asp |
| -t |  transform list | -t wordstranform.txt |
| -d |  domain to add to transform list | -d https://another.site.com |
| -ssrf |  SSRF url | -ssrf https://zeta2.beeceptor.com/x |
| -o |  output list name/location(default to current directory) | -o ./dirs.txt |



## Observations:

* dirsearch.py (dirsearch -u http://localhost/FUZZ -w ./words -e php)
    wordlist = 
    ```
    a0.txt
    /a.txt
    //a1.txt
    b.txt
    c.%EXT%
    d
    e f

    ```
    > without newline at end the last request "e f" in this case will have a %20 at end so "e f "

    Disearch requests ->
    ```
    site.com/a0.txt
    site.com/a.txt
    site.com//a1.txt
    site.com/b.txt
    site.com/c.php
    site.com/d
    site.com/e%20f
    ```

* ffuf (ffuf -u http://localhost/FUZZ -w ./words)
    wordlist = 
    ```
    a0.txt
    /a.txt
    //a1.txt
    b.txt
    c.%EXT%
    d
    e f
    h.%ok%
    ```

    ffuf requests->
    ```
    site.com/a0.txt
    site.com//a.txt
    site.com///a1.txt
    site.com/b.txt
    c.%EXT. <---  ERROR'd. ffuf tries to decode or something %EX error as its no valid
    site.com/d
    site.com/e%20f
    ```

* Output
    * Ffuf should have all the words URL-Encoded ( url.Parse("/test/some\"%22") then val.EscapedPath() )