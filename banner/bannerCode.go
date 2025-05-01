package banner

import "fmt"

func PrintBanner() {  // Changed to uppercase to export the function
	cyan := "\033[1;36m"  // Bold Cyan
	yellow := "\033[1;33m" // Bold Yellow
	green := "\033[1;32m"  // Bold Green
	blue := "\033[1;34m"   // Bold Blue
	reset := "\033[0m"

	fmt.Printf(cyan + `
 _______           ______   _______  _______  _______  _______ 
(  ____ \|\     /|(  ___ \ (  ____ \(  ___  )(  ____ )(  ____ \
| (    \/| )   ( || (   ) )| (    \/| (   ) || (    )|| (    \/
| (_____ | |   | || (__/ / | (__    | |   | || (____)|| (_____ 
(_____  )| |   | ||  __ (  |  __)   | |   | ||     __)(_____  )
      ) || |   | || (  \ \ | (      | |   | || (\ (         ) |
/\____) || (___) || )___) )| )      | (___) || ) \ \__/\____) |
\_______)(_______)|/ \___/ |/       (_______)|/   \__/\_______)
` + reset + "\n")

	fmt.Printf(yellow+"Version:   "+green+"v0.1\n"+reset)
	fmt.Printf(yellow+"Developer: "+green+"0xS22d\n"+reset)
	fmt.Printf(yellow+"Website:   "+blue+"https://saad-ayady.github.io/myWEBSITE/\n\n"+reset)
	fmt.Println(yellow + "------------------------------------------------" + reset)
}