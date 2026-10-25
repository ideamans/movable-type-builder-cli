# アーカイブ・日付

アーカイブページと日付の出力。

現行のタグ 29 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTArchiveCount — FUNCTION
いまのアーカイブに含まれる記事の数

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/archivecount.html)

### MTArchiveDate — FUNCTION
日付アーカイブの期間の始まりの日付

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/archivedate.html)

### MTArchiveDateEnd — FUNCTION
日付アーカイブの期間の終わりの日付

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/archivedateend.html)

### MTArchiveFile — FUNCTION
いまのアーカイブの出力ファイル名

属性: `extension="0 | 1"` `separator="-"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/archivefile.html)

### MTArchiveLink — FUNCTION
いまのアーカイブの絶対 URL

属性: `archive_type="archive_type_foo"` `with_index="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/archivelink.html)

### MTArchiveList — BLOCK
アーカイブを並べるループ（archive_type で種類を選ぶ）

属性: `archive_type="archive_type_foo"` `type="archive_type_foo"` `sort_order="ascend | descend"` `lastn="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/archivelist.html)

### MTArchiveListFooter — BLOCK
MTArchiveList の最後の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/archivelistfooter.html)

### MTArchiveListHeader — BLOCK
MTArchiveList の最初の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/archivelistheader.html)

### MTArchiveNext — BLOCK
次のアーカイブを文脈にする（中の MTArchiveTitle などが次を指す）

属性: `archive_type="archive_type_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/archivenext.html)

### MTArchivePrevious — BLOCK
前のアーカイブを文脈にする

属性: `archive_type="archive_type_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/archiveprevious.html)

### MTArchiveTitle — FUNCTION
いまのアーカイブの題（月別なら年月、カテゴリならカテゴリ名など）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/archivetitle.html)

### MTArchiveType — FUNCTION
アーカイブの種類の識別子（Monthly など）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/archivetype.html)

### MTArchiveTypeLabel — FUNCTION
アーカイブの種類の表示名（ブログの言語で）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/archivetypelabel.html)

### MTArchives — BLOCK
サイトが持つアーカイブの種類を並べるループ（サイトマップ向け）

属性: `archive_type="archive_type_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/archives.html)

### MTCalendar — BLOCK
1か月分のカレンダーを組むループ（month で月を選ぶ）

属性: `category="category_foo"` `month="this | last | YYYYMM"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/calendar.html)

### MTCalendarCellNumber — FUNCTION
カレンダーのセルの番号（左上から 1）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/calendarcellnumber.html)

### MTCalendarDate — FUNCTION
カレンダーのセルの日時（format で書式指定）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/calendardate.html)

### MTCalendarDay — FUNCTION
カレンダーのセルの日

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/calendarday.html)

### MTCalendarIfBlank — BLOCK
カレンダーの空のセルで中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/calendarifblank.html)

### MTCalendarIfContents — BLOCK
MTContentCalendar の中で、その日にコンテンツデータがあれば出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/calendarifcontents.html)

### MTCalendarIfEntries — BLOCK
その日に記事があれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/calendarifentries.html)

### MTCalendarIfNoContents — BLOCK
MTContentCalendar の中で、その日にコンテンツデータが無ければ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/calendarifnocontents.html)

### MTCalendarIfNoEntries — BLOCK
その日に記事が無ければ中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/calendarifnoentries.html)

### MTCalendarIfToday — BLOCK
今日のセルで中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/calendariftoday.html)

### MTCalendarWeekFooter — BLOCK
週の終わりで中身を出す（</tr> など）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/calendarweekfooter.html)

### MTCalendarWeekHeader — BLOCK
週の始まりで中身を出す（<tr> など）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/calendarweekheader.html)

### MTDate — FUNCTION
再構築した日時（format で書式指定）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/date.html)

### MTDateFooter — BLOCK
MTEntries の中で、次の記事と日付が変わる直前に出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/datefooter.html)

### MTDateHeader — BLOCK
MTEntries の中で、前の記事と日付が変わった直後に出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/dateheader.html)
