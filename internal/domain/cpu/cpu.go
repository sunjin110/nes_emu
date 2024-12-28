package cpu

import "github.com/sunjin110/nes_emu/internal/domain/memory"

const memorySize = 16 * 1024 // 16KB

// CPU document: https://www.nesdev.org/wiki/CPU
type CPU struct {
	memory   memory.Memory
	register Register
}

/**
# cpu実行順序
## fetch
- プログラムカウンタ(PC)が指している場所のROMから命令を読み込む。
命令によっては引数(オペランド)があることもあり、その場合はオペランドも読み込む

- この時、次のfetchのために次の命令を指すようにプログラムカウンタ(PCの値を更新する)

## decode
- ROMから読み込んだ命令の内容を解釈する

## execute
- 命令ごとに決められた演算を行う。これにより、レジスタの値やRAMに保存されている値が更新される。
*/

/**
CPU命令割り込み
RESET: 起動時とリセットボタンが押された時
NMI: ハードウェア割り込み。PPUが描画完了したことをCPUに知らせる時に使用
IRQ: APUのフレームシーケンサが発生させ利割り込み
BRK: BRK命令を実行した時に発生するもの
*/

/**
アドレッシングモード
命令を実行する時に引数を指定する方法が13種類ある、それをアドレッシングモードと呼ぶ
document: https://www.nesdev.org/wiki/CPU_addressing_modes

- Implied: 引数なし
- Accumulator: Aレジスタを利用する
- Immediate: 定数を指定する(8bitまで)
- Zeropage: アドレスを指定
- Zeropage, X: 配列などを渡せる
- Zeropage, Y: Zeropage, Xと同じ、レジスタの利用する箇所が違う
- Relative: 分岐命令で利用されるやつ
- Absolute: 変数に入っているものを指定するやつ
- Absolute, X: Zeropage, Xと同じだが変数を指定
- Absolute, Y: Absolute, Xと同じ、レジスタの利用する箇所が違う
- (Indirect): 引数として16bitの値を利用する、括弧は参照はずしなのでIM16をアドレスとしてみた時のIM16番地にあるアタういを表す
- (Indirect, X): 8bitのアドレス(IM8 + X)に格納されているアドレスを操作対象とする -> 配列のN番目を取得する的な
- (Indirect), Y: 上記とは違う
*/

/**
(Indirect, X)と(Indirect), Yの違い

特徴	(Indirect, X)	(Indirect), Y
参照の段階数	1段階（ゼロページから直接間接アドレスを取得）	2段階（ゼロページ参照 + オフセット加算）
操作順序	即値 + X → ゼロページ参照	即値 → ゼロページ参照 → + Y
主な用途	ポインタテーブル（複数ポインタの中の1つを選ぶ）	配列やデータブロック（ベースアドレスにオフセットを加える）
例えるなら	「テーブルから直接1つの値を取る」	「ベースアドレスにオフセットを足して次の値を取る」

# (Indirect), Yが必要な理由
base addressを相対的に決めてデータアクセスができるため便利、例えばスプライトデータが$4000からだとして、そのデータを取得する場合
IndirectXだけの場合は常に$4000を意識しながらデータをアクセスする必要がある
*/
