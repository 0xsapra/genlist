package main

import (
	"fmt"
	"flag"
	"strings"
	"net/url"
	"errors"
	"os"
	"bufio"
	// "path"

	tld "github.com/jpillora/go-tld"
)

type arrayFlags []string
func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var Wordlists arrayFlags // input wordlist's
var Transformlists arrayFlags // wordlist's that should be transformed
var Extensions string // extensions comma seperated without .
var Domain string // domain name (schema://domain/path )
var SSRF string // SSRF url
var OutputFormat string // -of enum(dirsearch/ffuf) output format
var Output string // -o /tmp/a.txt .. Default to ./Domain.txt

var AllWords []string
var ExtensionList []string

var CorrectOutputFormats = map[string]bool {
	"ffuf" : true,
	"dirsearch" : true,
};

var TRANSFORM_WORDS = map[string]string {}

func main () {
	
	// Normal wordlist
	flag.Var(&Wordlists, "w", "Wordlist")
	// Transform wordlist
	flag.Var(&Transformlists, "t", "Transform Wordlist")
	// Extensions comma seperated
	flag.StringVar(&Extensions, "e", "php,asp,jsp,js,txt,zip" ,"List of extensions. Defaults to php,asp,jsp,js,txt,zip")
	// Domain (schema<http?s should be there>)
	flag.StringVar(&Domain, "d", "", "Domain name with http/https")
	// SSRF
	flag.StringVar(&SSRF, "ssrf", "", "Domain to send ssrf proble<with http/https>")
	// output format enum[dirsearch/ffuf]. Default ffuf
	flag.StringVar(&OutputFormat, "of", "ffuf", "output format enum[dirsearch/ffuf]. Default ffuf")
	// output
	flag.StringVar(&Output, "o", "", "output file(default to ./domain.txt")

	flag.Parse()

	if len(Wordlists) == 0  && len(Transformlists) == 0{
		fmt.Println("[-] Both wordlist and Transformlists cannot be empty. required -w or -t or both")
		return;
	}

	ExtensionList = strings.Split(Extensions, ",")
	
	if _, found := CorrectOutputFormats[OutputFormat]; found == false {
		fmt.Println("[-] output format can be either dirsearch or ffuf.", OutputFormat)
		return;
	}
	
	if _, err := validateDomain(Domain); err != nil {
		fmt.Println(err, "For Domain name")
		return;
	}
	if SSRF == "" {
		fmt.Println("[+] SSRF name not present. Using https://zeta2.free.beeceptor.com")
		SSRF = "https://zeta2.free.beeceptor.com"
	}
	if _, err := validateDomain(SSRF); err != nil {
		fmt.Println(err, "For SSRF name")
		return;
	}
	if Output == "" {
		Output = "./" + snakeCaseDomain(Domain) + ".dirlist.txt"
		fmt.Println("[+] No output file given so outputing to ", Output)
	}

	genList(Wordlists, Transformlists, ExtensionList, Domain, SSRF, Output, OutputFormat)	
	

}


func genList(wordlists, transformlist, extensions []string, domain, ssrf, output, oformat string) {
	
	var words []string
	var trasnformlist []string
	var allWords []string
	var subdomainSuffex string
	tldUrl, _ := tld.Parse(domain)

	if tldUrl.Subdomain != "" {
		subdomainSuffex = tldUrl.Subdomain + "."
	} else {
		subdomainSuffex = ""
	}

	TRANSFORM_WORDS["{domain}"] = tldUrl.Domain
	TRANSFORM_WORDS["{url}"] = tldUrl.Domain + "." + tldUrl.TLD
	TRANSFORM_WORDS["{fullurl}"] = subdomainSuffex + tldUrl.Domain + "." + tldUrl.TLD
	TRANSFORM_WORDS["{_url}"] = snakeCaseDomain(TRANSFORM_WORDS["{url}"])
	TRANSFORM_WORDS["{_fullurl}"] = snakeCaseDomain(TRANSFORM_WORDS["{fullurl}"])
	
	ssrfUrl, _ := url.Parse(ssrf)
	ssrfUrl.Scheme = ""
	TRANSFORM_WORDS["{ssrf_here}"] = ssrfUrl.String()
	TRANSFORM_WORDS["{ssrf_here_2}"] = ssrfUrl.String()[2:]
	
	if len(wordlists) != 0 {
		for _, wordlist := range(wordlists) {
			if _words, err := readWordsFromFile(wordlist); err == nil {
				words = append(words, _words...)
			} else {
				fmt.Println(err)
			}
		}
	}

	if len(transformlist) != 0 {
		for _, wordlist := range(transformlist) {
			if _words, err := readWordsFromFile(wordlist); err == nil {
				trasnformlist = append(trasnformlist, _words...)
			} else {
				fmt.Println(err)
			}
		}
	}
	
	if oformat == "ffuf" {
		
		allWords = parseWordlistFFUF(words, extensions)
		allWords = append(
			parseTransformListFFUF(trasnformlist, extensions), 
			allWords...
		)

		for i, _ := range(allWords) {
			if allWords[i][0] == '/' {
				allWords[i] = allWords[i][1:]
			}
		}

	} else if oformat == "dirsearch" {
		allWords = append(
			parseTransformListFFUF(trasnformlist, extensions), 
			words..., 
		)
	} else {
		panic("How DID You even reached here")
	}

	if done, err := writeToFile(allWords, output); done == true {
		return;
	} else {
		fmt.Println("[-]",err)
		return;
	}
 
}

func readWordsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}

func parseWordlistFFUF(words []string, extensions []string) []string {
	var transformedWords = make([]string, len(words) )
	var extensionIdentifier = "%EXT%"

	for _, word := range(words) {
		if strings.Contains(word, extensionIdentifier) {
			var newdirs = make([]string, len(extensions))

			for i, extension := range(extensions) {
				newdirs[i] = strings.ReplaceAll(word, extensionIdentifier, extension)
			}

			transformedWords = append(transformedWords, newdirs...)
		} else {
			transformedWords = append(transformedWords, word)
		}

	}

	return transformedWords;
}

func parseTransformListFFUF(words, extensions []string) ([]string) {
	var transformedWords = make( []string, len(words) )

	for i, word := range(words) {
		transformedWords[i] = word

		for tranformIdentifier, value := range(TRANSFORM_WORDS) {

			if strings.Contains(transformedWords[i], tranformIdentifier) {
				transformedWords[i] = strings.ReplaceAll(transformedWords[i], tranformIdentifier, value)
			}

		}
	}

	return transformedWords;
}


func tranformWord(word string, extensions []string) []string {
	
	// {domain}  = site
	// {url}     = site.com      	// without subdomain
	// {fullurl} = www.site.com  	// with subdomain
	// {_fullurl} = www_site_com  	// with subdomain
	// {_url}    = site_com      	// without subdoamin site_com
	// {ssrf_here} = //zzzzz.beeceptop.com  // any site u wanna check ssrf for
	// {ssrf_here_2} = zzzzz.beeceptop.com  // without the "//""
	
	// var words = make([]string)
	
	// for i, extension := range(extensions) {
	// 	words[i] = strings.ReplaceAll(word, "%EXT%", extension)
	// }

	// if (domain == "" && ssrf == "") {
	// 	return words;
	// }
	// return words
	return []string{"xx"}
}

// check if domain has proper schema and properly formatted i.e schema://domain.tld
func validateDomain(domain string) (bool, error) {
	u, err := url.Parse(domain)

	if err != nil {
		return false, err
	}

	if u.Scheme == "" {
		return false, errors.New("Schema missing. Required https:// OR http://")
	}
	if (u.Scheme != "http" && u.Scheme != "https") {
		return false, errors.New("Schema Can only be http or https and given" + u.Scheme)
	}

	return true, nil
}

// change a.b.c to a_b_c
func snakeCaseDomain(domain string) string {
	u, err := url.Parse(domain)

	if err != nil || u.Host == "" {
		return strings.Join(strings.Split(domain, ".")[:], "_")
	} 

	return strings.Join(strings.Split(u.Host, ".")[:], "_")
}


func writeToFile(words []string, filePath string) (bool, error) {

	f, err := os.Create(filePath)

	if err != nil {
        return false, err
    }
	defer f.Close()

	for _, word := range words {

        _, err := f.WriteString(word + "\n")

        if err != nil {
            continue
        }
    }
	return true, nil

}