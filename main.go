package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"text/template"
	"time"
)

// OptMain is main処理格納先
type OptMain struct {
	Sorted  bool
	GetItem bool
	Item1   string
	Item2   string
}

// Page is templateページ全部
type Page struct {
	Title       string
	Description string
	ValA        string
	ValB        string
	ItemView    string
	SortView    string
	ResultView  string
	Data        []ItemInfo
}

// ポインタ参照にしないとエラーになる
var inputitems []ItemInfo
var leftitem string
var rigthitem string
var itemtheme string
var inputdescription string
var selectcount int
var findkey map[string]int

// ItemInfo is 実施結果
type ItemInfo struct {
	Seqno  int
	Rankno int
	Item   string
	Point  int
}


// main is 開始処理
func main() {
	fs := http.FileServer(http.Dir("./resource"))

	// 初期画面
	http.HandleFunc("/init", initHandler)
	// はじめるボタン押下
	http.HandleFunc("/start", startHandler)
	// 好きな方を選択
	http.HandleFunc("/sel", viewHandler)
	http.Handle("/data/", http.StripPrefix("/data/", fs))

	http.ListenAndServe(":9999", nil)

	selectcount = 0

}

// initHandler is  "/"の時のページ。初期ページ。
func initHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("initHandler-go!")
	// テンプレート用のファイルを読み込む
	tpl, err := template.ParseFiles("view/index.html")
	Check(err)

	initret := make([]ItemInfo, 0)
	page := Page{"ソートしたい内容を入力してください.", itemtheme, "a", "b", "show-box", "hide-box", "hide-box", initret}
	err = tpl.Execute(w, page)

	Check(err)
}

// startHandler is はじめる選択時
func startHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("startHandler-go!")

	// 初回は入力された内容を読み込んで配列にする。
	inputval := r.FormValue("inputitems")
	inputval2 := r.FormValue("inputtheme")
	inputval3 := r.FormValue("inputdescription")

	if len(inputval2) > 0 {
		itemtheme = inputval2 + "ソート"
	}
	if len(inputval3) > 0 {
		inputdescription = inputval3
	}

	// 改行が含まれているかどうか
	if strings.Index(inputval, "\r\n") > 0 {

		findwords := strings.Split(inputval, "\r\n")
		// 配列の初期サイズをセット
		inputitems = make([]ItemInfo, 0)
		var assignitem ItemInfo
		// チェック済みかの判定用配列
		findkey = make(map[string]int, 99)
		// １行ずつ検索する
		for index := 0; index < len(findwords); index++ {
			keyval := findwords[index]

			// ブランク行は無視する
			if len(strings.TrimSpace(keyval)) < 1 {

				continue
			}
			fmt.Printf("配列にセット[%v]=%v \n", index, keyval)
			// ソート用配列に詰める
			assignitem.Item = keyval
			assignitem.Seqno = index + 1
			assignitem.Point = 0
			assignitem.Rankno = 0

			switch index {
			case 0:
				leftitem = keyval
			case 1:
				rigthitem = keyval

			}
			// 配列に追加する
			inputitems = append(inputitems, assignitem)

		}

	} else {
		fmt.Println("改行なし")
	}
	findkey[leftitem+"v.s."+rigthitem] = 1

	// テンプレート用のファイルを読み込む
	tpl, err := template.ParseFiles("view/index.html")
	Check(err)

	// まだ使わない
	initret := make([]ItemInfo, 0)
	// 探索済みであればチェックする

	page := Page{"開始", inputdescription, leftitem, rigthitem, "hide-box", "show-box", "hide-box", initret}
	err = tpl.Execute(w, page)

}

// viewHandler is アイテム選択時
func viewHandler(w http.ResponseWriter, r *http.Request) {

	// // // まだ使わない
	// rlt := make([]ItemInfo, 0)

	// 2回目以降、値が選ばれているので取得し値をセット
	valleft := r.PostFormValue("valA")
	valrigth := r.PostFormValue("valB")
	valeq := r.PostFormValue("eq")
	selectcount++
	// 入力値のセット
	// 左右どちらも選択されなかったとき
	if len(valeq) > 0 {
		// 同値は何もしない
	} else {
		// ポイント加算のための添え字取得
		findval := func(itemname string) int {
			var ret int
			ret = -1
			for index := 0; index < len(inputitems); index++ {
				if inputitems[index].Item == itemname {
					ret = index
					break
				}
			}
			return ret
		}
		// 名前を添え字に変換
		x := findval(leftitem)
		y := findval(rigthitem)
		// 選ばれたものは加算する
		if len(valleft) > 0 {

			inputitems[x].Point++
			inputitems[y].Point--
		}

		if len(valrigth) > 0 {
			inputitems[x].Point--
			inputitems[y].Point++
		}

	}

	// 次回の準備
	leftitem = ""
	rigthitem = ""

	// メインロジック
	o := execSort()
	leftitem = o.Item1
	rigthitem = o.Item2
	// 次のアイテムを取得
	// テンプレート用のファイルを読み込む
	tpl, err := template.ParseFiles("view/index.html")
	Check(err)

	// ソート完了の場合
	if o.Sorted {

		// 所持pt順に並べる
		sort.Slice(inputitems, func(i, j int) bool {
			return inputitems[i].Point > inputitems[j].Point
		})

		rank := 1
		for index := 0; index < len(inputitems); index++ {

			if index > 0 {
				if inputitems[index].Point == inputitems[index-1].Point {
					// 1つ前と同率であれば順位変動なしとする
				} else {
					rank++
				}
			}
			inputitems[index].Rankno = rank
		}

		page := Page{itemtheme, inputdescription, "", "", "hide-box", "hide-box", "cols-box", inputitems}
		err = tpl.Execute(w, page)
		Check(err)
	} else {
		pagetitle := fmt.Sprintf("%v [%v回目]", itemtheme, selectcount)
		page := Page{pagetitle, inputdescription, leftitem, rigthitem, "hide-box", "show-box", "hide-box", inputitems}
		err = tpl.Execute(w, page)
		Check(err)
	}

}

// 参考
// http://www.ics.kagoshima-u.ac.jp/~fuchida/edu/algorithm/sort-algorithm/

// execSort is ソートのメイン処理
func execSort() OptMain {
	var obj OptMain

	//ソートさせる
	// 次のアイテムを取得する
	nextitem := func(item string, itemseq int) OptMain {

		for index := 0; index < len(inputitems); index++ {
			k := inputitems[index].Item
			seq := inputitems[index].Seqno

			// 違うキー同士の場合
			if item != k {

				// INDEXが小さい方を前にもってくるため比較
				if seq < itemseq {

					obj.Item1 = k
					obj.Item2 = item

				} else {

					obj.Item1 = item
					obj.Item2 = k

				}
				// 存在チェック
				_, flg := findkey[obj.Item1+"v.s."+obj.Item2]
				if flg {
					// 存在している場合
					// チェック済みなので何もしない
					continue

				} else {
					// いなかったとき
					findkey[obj.Item1+"v.s."+obj.Item2] = 1
					obj.GetItem = true
					return obj
				}
			}

		}

		return obj
	}

	// 要素数分ループ
	for index := len(inputitems); index > 0; index-- {
		index--
		// 次の値を取得する
		obj = nextitem(inputitems[index].Item, inputitems[index].Seqno)

		if obj.GetItem {
			// 次の表示アイテムが取得できた場合
			obj.Sorted = false
			break
		} else {
			// ソート終わった場合
			obj.Sorted = true
		}
	}

	// 次の表示アイテム取得処理＆結果表示用

	return obj
}

// goNextStepForRandom is 次の要素をランダムで取得するためのメソッド。
func goNextStepForRandom() OptMain {
	var obj OptMain

	// 要素数分ループ
	for index := 0; index < len(inputitems); index++ {
		// Seed作成
		var inf64 int64
		inf64 = int64(index + 1)
		r := rand.New(rand.NewSource(67))
		r.Seed(time.Now().UnixNano() + inf64)
		newidx := r.Intn(len(inputitems) - 1)

		inf64 = int64(newidx + 1)
		r.Seed(time.Now().UnixNano() + inf64)
		newidx2 := r.Intn(len(inputitems) - 1)

		// アイテム取得
		k := inputitems[newidx].Item
		seq := inputitems[newidx].Seqno

		// アイテム取得２
		item := inputitems[newidx2].Item
		itemseq := inputitems[newidx2].Seqno

		// 違うキー同士の場合
		if item != k {

			// INDEXが小さい方を前にもってくるため比較
			if seq < itemseq {

				obj.Item1 = k
				obj.Item2 = item

			} else {

				obj.Item1 = item
				obj.Item2 = k

			}
			// 存在チェック
			_, flg := findkey[obj.Item1+"v.s."+obj.Item2]
			if flg {
				// 存在している場合
				// チェック済みなので何もしない
				continue

			} else {
				// いなかったとき
				findkey[obj.Item1+"v.s."+obj.Item2] = 1
				obj.GetItem = true
				obj.Sorted = false
				break
			}
		}

	}
	return obj
}

// Check is エラーチェック
func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
