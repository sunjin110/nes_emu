package cpu

type Register struct {
	a  byte    // Accumulator 命令の計算結果の格納
	x  byte    // 特定のアドレッシングモード (後述) でインデックスとして使われます。 INX 命令と組み合わせてループのカウンタとしても使われる様子？
	y  byte    // Xと同様
	pc [2]byte // Program Counter // CPUが次に実行すべき命令のアドレスを保持する
	sp [2]byte // Stack Pointer // スタックの先頭のアドレスを保持します
	// P Processor Status // ステータスレジスタ。各ビットが意味を持つ、
	//  file:///Users/sunjin/Downloads/LayerWalker.pdf
	p byte
}

// TODO ステータスレジスタの更新と読み取り
// file:///Users/sunjin/Downloads/LayerWalker.pdf
