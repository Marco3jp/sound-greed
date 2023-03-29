# sound-greed

## 注意点
- yt-dlp周り
  - パッケージ管理システムによっては一部の依存関係がoptionalになっていることがあります
  - 多分ffmpegは入れる必要があって、あとなんかエラー出たらいい感じに直してください
- 設定を変えたらビルドし直さないと反映されないです
  - なぜか過去の私が設定ファイルもembedにしたため
  - 逆に言えば設定書いて固めてしまえばバイナリだけで動きます

## 参考
- sound-greed.serviceを使うといい感じにsystemdに任せられます
  - アプリケーションのほうは設定書いてビルドする
    - 普通に実行させて動作確認するのあり（特に保存先とか意図通りに動くか）
  - `ln -s /path/to/sound-greed/main ~/bin/sound-greed` する（しなくてもいいけど）
  - sound-greed.serviceのExecStartを調整
  - `/home/username/.config/systemd/user/sound-greed.service` に移動
    - ここに置くとユーザースコープで扱える
  - `systemctl --user enable --now sound-greed.service`
    - 移動させたあとになんか読み込むコマンド打たないとだめかも
