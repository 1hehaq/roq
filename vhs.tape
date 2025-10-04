Output roq.gif

Set FontFamily "SF Mono"
Set WindowBar Colorful
Set BorderRadius 15
Set Padding 20
Set Margin 20
Set MarginFill "#0c1117"

Type "roq -h"
Enter
Type "roq -list | wc -l"
Enter
Sleep 3s
Type "clear"
Enter
Type "roq -s github -k ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
Enter
Sleep 1s
Type "roq -s shodan -k xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx -json | jq"
Sleep 1s
Enter
Sleep 3s
Ctrl+W
Ctrl+W