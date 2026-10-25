package main

import "regexp"

// groupRules はタグ名を機能別のまとまりに割り当てます。
//
// 上から順に当て、最初に当たったものが勝ちます。**順番に意味がある**
// （folder を category より先に見ないと HasSubFolders が category に落ちる）。
// movabletype.jp の索引は五十音順なので、機能で引けるようにするための分類です。
var groupRules = []struct {
	group string
	re    *regexp.Regexp
}{
	{"customfield", regexp.MustCompile(`(?i)^MTCustomField`)},
	{"content", regexp.MustCompile(`(?i)^MTContent`)},
	{"entry", regexp.MustCompile(`(?i)^MT(Entries|Entry|BlogEntryCount)`)},
	{"page", regexp.MustCompile(`(?i)^MT(Pages|Page[A-Z]|BlogPageCount)`)},
	{"folder", regexp.MustCompile(`(?i)^MT(Folder|SubFolder|TopLevelFolder|ParentFolder|Has(No)?(Parent|Sub)Folder|TopLevelParent)`)},
	// EntryCategor / PageCategor をここに書いても当たらない（entry / page が先）。
	// MTEntryCategory は記事の文脈で使うタグなので entry のままでよい。
	{"category", regexp.MustCompile(`(?i)^MT(Categor|SubCateg|TopLevelCateg|ParentCategor|ArchiveCategor|Has(No)?(Parent|Sub)Categor)`)},
	{"asset", regexp.MustCompile(`(?i)^MT(Asset|Image|Audio|Video|File)`)},
	{"tag", regexp.MustCompile(`(?i)^MT(Tag|EntryTag|PageTag|AssetTag|ContentTag)`)},
	{"user", regexp.MustCompile(`(?i)^MT(Author|User|Signin|SignOut|PasswordValidation|Profile|OtherProfile)`)},
	{"comment", regexp.MustCompile(`(?i)^MT(Comment|Ping|Trackback|Feedback|Captcha|Registration|Remember|TypeKey)`)},
	{"archive", regexp.MustCompile(`(?i)^MT(Archive|Calendar|Date)`)},
	{"search", regexp.MustCompile(`(?i)^MT(Search|Results|NoSearch)`)},
	{"site", regexp.MustCompile(`(?i)^MT(Site|Blog|Website|ChildSites|ParentSite|OtherBlog|MultiBlog)`)},
	{"pagination", regexp.MustCompile(`(?i)^MT(NextLink|PreviousLink|CurrentPage|MaxResults|Pagination|TotalPages)`)},
	{"control", regexp.MustCompile(`(?i)^MT(If|Unless|Else|ElseIf|Loop|For|SetVar|SetHashVar|Var|Get|Include|Ignore|Block|Sub|Template|Widget|App|Section|Assign|Index)`)},
	{"format", regexp.MustCompile(`(?i)^MT(TextFormat|Textile|Markdown|SmartyPants|Text[A-Z]|Convert|RemoveHTML|SanitizeSpec)`)},
	{"system", regexp.MustCompile(`(?i)^MT(CGI|AdminCGI|AdminScript|Static|Version|Product|Config|PHP|Error|Http|Link|Script|Support|Publish|Rebuild|Canonical|DataAPI|DefaultLanguage|Stats|BuildTemplateID)`)},
}

// groupOf はタグ名から所属する群を返します。どれにも当たらなければ "other"。
func groupOf(name string) string {
	for _, r := range groupRules {
		if r.re.MatchString(name) {
			return r.group
		}
	}
	return "other"
}
