package main

import (
	"fmt"
	"github.com/abadojack/whatlanggo"
	"strings"
)

func checkLang(pTagText string, siteTitle string, compareLang string) bool {
	var content string
	enoughLenChunk := 400
	lenText := len(pTagText)
	if lenText == 0 {
		content = siteTitle
	} else {
		content = pTagText
	}

	var textChunk string

	// create text chunk to detect language
	if len(content) >= enoughLenChunk {
		slicePos := 0
		nChunks := 5
		chunksLenPosGap := lenText / nChunks
		lenSubChunk := enoughLenChunk / nChunks
		for i := 0; i < nChunks; i++ {
			textChunk = textChunk + content[slicePos:slicePos+lenSubChunk]
			slicePos += chunksLenPosGap
		}
	} else {
		textChunk = content
	}

	chunkLang := whatlanggo.DetectLang(textChunk)
	fmt.Println("whatlanggo.Langs[chunkLang] -- ", whatlanggo.Langs[chunkLang])
	fmt.Println("textChunk -- ", textChunk)
	if whatlanggo.Langs[chunkLang] == compareLang {
		return true
	}
	return false
}

func findCharPos(str string, char string, numOccurrences int, reversed bool) int {
	lenString := len(str)

	if !reversed {
		for i := 0; i < lenString; i++ {
			if str[i:i+1] == char {
				numOccurrences--
				if numOccurrences == 0 {
					return i
				}
			}
		}
	} else {
		for i := lenString - 1; i >= 0; i-- {
			if str[i:i+1] == char {
				numOccurrences--
				if numOccurrences == 0 {
					return i
				}
			}
		}
	}

	return -1
}

func getAllSiteUkrLinks(linkToChange string) []string {
	// supported links changes -- www.example.com/uk/, www.example.com/ua/, www.example.de,
	// de.example.com, www.example.com/about/?lang=en,
	// www.example.com.ua, uk-ua.example.com;
	// taken from here -- https://wpml.org/documentation/getting-started-guide/language-setup/language-url-options/

	var arrayLinks []string

	lenLink := len(linkToChange)
	if linkToChange[lenLink-1:lenLink] != "/" {
		fmt.Println(linkToChange[lenLink-1 : lenLink])
		linkToChange += "/"
	}

	arrayLinks = append(arrayLinks, linkToChange)

	if strings.Contains(linkToChange, "https://") {
		// check en. (or other language) in links similar to https://en.wikipedia.org/wiki/
		if linkToChange[10:11] == "." {
			linkToChange = linkToChange[:8] + linkToChange[11:]
		}
	}

	// add new links change
	arrayLinks = append(arrayLinks, linkToChange+"uk/")
	arrayLinks = append(arrayLinks, linkToChange+"ua/")
	arrayLinks = append(arrayLinks, linkToChange+"ukr/")

	// add new link change
	arrayLinks = append(arrayLinks, linkToChange+"?lang=uk/")

	if !strings.Contains(linkToChange, ".ua/") {
		// add new link change
		endSubstringPos := findCharPos(linkToChange, "/", 3, false)
		startDomainPos := findCharPos(linkToChange[:endSubstringPos], ".", 1, true)
		if startDomainPos != -1 {
			arrayLinks = append(arrayLinks, linkToChange[:startDomainPos]+".ua/")
		}

		// add new link change
		wwwPos := strings.Index(linkToChange, "www.")
		if wwwPos != -1 {
			arrayLinks = append(arrayLinks, linkToChange[:wwwPos]+"uk."+linkToChange[wwwPos+4:])
			arrayLinks = append(arrayLinks, linkToChange[:wwwPos]+"ua."+linkToChange[wwwPos+4:])
		} else {
			arrayLinks = append(arrayLinks, linkToChange[:8]+"ua."+linkToChange[8:])
			arrayLinks = append(arrayLinks, linkToChange[:8]+"uk."+linkToChange[8:])
		}

		// add new link change
		endDomainPos := strings.Index(linkToChange, ".com/")
		if endDomainPos != -1 {
			arrayLinks = append(arrayLinks, linkToChange[:endDomainPos+4]+".ua/"+linkToChange[endDomainPos+5:])
		}

		// add new link change
		if wwwPos != -1 {
			arrayLinks = append(arrayLinks, linkToChange[:wwwPos]+"uk-ua."+linkToChange[wwwPos+4:])
		}
	}

	return arrayLinks
}
