package cpu

// Mnemonic 命令
// document: https://www.nesdev.org/opcodes.txt
type Mnemonic int

const (
	LDA Mnemonic = iota // Aレジスタにロードする
	LDX                 // Xレジスタにロードする
	LDY                 // Yレジスタにロードする
	STA                 // Aレジスタをストアする
	STX                 // Xレジスタをストアする
	STY                 // Yレジスタをストアする
	TXA                 // XレジスタをAレジスタにコピーする
	TYA                 // YレジスタをAレジスタにコピーする
	TXS                 // XレジスタをSPレジスタにコピーする
	TAY                 // AレジスタをYレジスタにコピー
	TAX                 // AレジスタをXレジスタにコピー
	TSX                 // SPレジスタをXレジスタにコピーする
	PHP                 // フラグをプッシュする
	PLP                 // フラグをポップする
	PHA                 // Aレジスタをプッシュする
	PLA                 // Aレジスタをポップする
	ADC                 // Aレジスタに加算する
	SBC                 // Aレジスタから減算する
	CMP                 // Aレジスタと比較する
	CPX                 // Xレジスタと比較する
	CPY                 // Yレジスタと比較する
	AND                 // AレジスタとAND演算をする
	EOR                 // AレジスタとEX-OR演算をする
	ORA                 // AレジスタとOR演算をする
	BIT                 // AレジスタとAND比較をする
	ASL                 // 左シフト
	LSR                 // 右シフト
	ROL                 // 左ローテイト
	ROR                 // 右ローテイト
	INC                 // 1を加算する
	INX                 // Xレジスタに1を加算する
	INY                 // Yレジスタに1を加算する
	DEC                 // 1を減算する
	DEX                 // Xレジスタから1を減算する
	DEY                 // Yレジスタから1を減算する
	CLC                 // Cフラグをクリア
	CLI                 // Iフラグをクリア
	CLV                 // Vフラグをクリア
	CLD                 // Dフラグをクリア
	SEC                 // Cフラグをセット
	SEI                 // Iフラグをセット
	SED                 // Dフラグをセット
	NOP                 // 何もしない
	BRK                 // ソフトウエア割り込み
	JMP                 // ジャンプ
	JSR                 // サブルーチン呼び出し
	RTS                 // サブルーチンから復帰
	RTI                 // 割り込み処理から復帰
	BCC                 // Branch if Carry Clear: キャリーフラグがクリアされている場合に分岐します。	フラグ C が 0 の場合に分岐します。
	BCS                 // Branch if Carry Set: キャリーフラグがセットされている場合に分岐します。: フラグ C が 1 の場合に分岐します。
	BEQ                 // Branch if Equal: ゼロフラグがセットされている場合に分岐します。: フラグ Z が 1 の場合に分岐します。

	// unofficial
	ALR // (AND + LSR): 累積レジスタ (A) に指定値と AND 演算を行い、結果を右シフトします。
	ANC // (AND + Carry): A レジスタと指定値に AND を適用し、さらにキャリーフラグに影響を与えます。
	ARR // (AND + ROR): A レジスタと指定値に AND を適用した後に、右ローテート (ROR) を行います。
	AXS // (AND + Subtract): X レジスタと指定値に AND を適用し、A レジスタを使って減算します。
	LAX // (Load A and X): A レジスタと X レジスタに同じ値を同時にロードします。
	SAX // (Store A AND X): A レジスタと X レジスタの AND 結果をメモリに保存します。
	DCP // (Decrement + Compare): メモリの値をデクリメントし、結果を A レジスタと比較します。
	ISC // (Increment + SBC): メモリの値をインクリメントし、その結果を累積レジスタから減算します。
	RLA // (ROL + AND): メモリの値を左ローテート (ROL) した後、累積レジスタ (A) に AND を適用します。
	RRA // (ROR + ADC): メモリの値を右ローテート (ROR) した後、累積レジスタに加算します。
	SLO // (ASL + ORA): メモリの値を左シフト (ASL) した後、累積レジスタに OR 演算を行います
	SRE // (LSR + EOR): メモリの値を右シフト (LSR) した後、累積レジスタに排他的論理和 (EOR) を適用します。
	SKB // (Skip Byte): 次のバイトをスキップします (NOP に似ていますが、引数をスキップします)。
	IGN // (Ignore Byte): 引数を持つ NOP として動作します。
)
