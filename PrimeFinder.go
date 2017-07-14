/*
	 ____       _                _____ _           _
	|  _ \ _ __(_)_ __ ___   ___|  ___(_)_ __   __| | ___ _ __
	| |_) | '__| | '_ ` _ \ / _ \ |_  | | '_ \ / _` |/ _ \ '__|
	|  __/| |  | | | | | | |  __/  _| | | | | | (_| |  __/ |
	|_|   |_|  |_|_| |_| |_|\___|_|   |_|_| |_|\__,_|\___|_|

	PrimeFinder

	Copyright (c) 2017 Jacob O'Toole

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.

*/

package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var primes = make([]uint64, 0)
var newPrimes = make([]uint64, 0)
var current uint64 = 3
var finishAt uint64
var toFind int

var isCont bool

var saveFile string
var loc int64
var writing int
var continueWrite bool
var loadChunk int64

var line1 string
var line2 string
var line3 string

//See if a number is prime with previous knowledge
func isPrime(num uint64) bool {
	testTo := uint64(math.Ceil(math.Sqrt(float64(num))))
	for current := 0; primes[current] <= testTo && current < len(primes)-1; current++ {
		if num%primes[current] == 0 {
			return false
		}
	}
	return true
}

//'Brute' check a number
func hardIsPrime(num uint64) bool {
	if num == 2 {
		return true
	} else if num%2 == 0 || num == 1 {
		return false
	}
	testTo := uint64(math.Ceil(math.Sqrt(float64(num))))
	for current := uint64(3); current <= testTo; current += 2 {
		if num%current == 0 {
			return false
		}
	}
	return true
}

//Function to write the newly discovered primes
func save() {
	fmt.Print(line3 + "Saving to " + saveFile + "...")
	//Open the save file for appending
	var file *os.File
	var err error
	if continueWrite {
		file, err = os.OpenFile(saveFile, os.O_APPEND, 0660)
	} else {
		file, err = os.Create(saveFile)
	}
	if err != nil {
		//Create the file if it doesn't exist and rerun save before returning
		if os.IsNotExist(err) {
			file, err := os.Create(saveFile)
			if err != nil {
				panic(err)
			}
			file.Close()
			save()
			return
		}
		//But panic if we can't open the file for any other reason
		fmt.Println("Error opening file!")
		panic(err)
	}
	//If we are continuing to write, we need to add ", " to the end before we start writing the new primes
	//But if not, then the next time we are continuing to write
	if continueWrite {
		file.WriteAt([]byte(", "), loc)
		loc += 2
	} else {
		continueWrite = true
	}
	//Var to store the final write at the end - without the ", " at the end
	var toWriteF []byte
	//Write with either primes or new primes depending on how we are calculating the primes
	if !isCont {
		//Write all but the last prime
		for writing < len(primes)-1 {
			toWrite := []byte(strconv.FormatUint(primes[writing], 10) + ", ")
			file.WriteAt(toWrite, loc)
			loc += int64(len(toWrite))
			writing++
		}
		toWriteF = []byte(strconv.FormatUint(primes[len(primes)-1], 10))
	} else {
		//Write all but the last prime
		for writing < len(newPrimes)-1 {
			toWrite := []byte(strconv.FormatUint(newPrimes[writing], 10) + ", ")
			file.WriteAt(toWrite, loc)
			loc += int64(len(toWrite))
			writing++
		}
		toWriteF = []byte(strconv.FormatUint(newPrimes[len(newPrimes)-1], 10))
	}
	//Do the final write
	file.WriteAt(toWriteF, loc)
	loc += int64(len(toWriteF))
	writing++
	//Close the file
	err = file.Close()
	if err != nil {
		panic(err)
	}
	t := time.Now()
	fmt.Println(line3 + "Last saved at: " + t.Format("2 Jan at 15:04") + " to " + saveFile)
}

//Returns the minimum int64
func minInt64(x, y int64) int64 {
	if x > y {
		return y
	} else {
		return x
	}
}

//Load primes from loadFile and return the amount of primes loaded
func load(loadFile, beforelog string) (loadUpTo uint64) {
	//Load upto the square root of the largest prime unless there is a number to find and we are going to infinity
	if loadFile == saveFile && (toFind == math.MaxInt32 || finishAt != math.MaxUint64) {
		loadUpTo = uint64(math.Ceil(math.Sqrt(float64(finishAt))))
		fmt.Print(beforelog + "Loading primes from " + loadFile + " up to " + strconv.FormatUint(loadUpTo, 10) + "...")
	} else {
		loadUpTo = math.MaxUint64
		fmt.Print(beforelog + "Loading all primes from " + loadFile + "...")
	}
	//Open the file
	filein, err := os.Open(loadFile)
	if err != nil {
		panic(err)
	}
	//Get size
	stats, err := filein.Stat()
	if err != nil {
		panic(err)
	}
	size := stats.Size()
	//Read and parse file
	var loading int64
	var lastVal string
	for ; loading < size; loading += loadChunk {
		//Run GC to keep mem usage low
		runtime.GC()
		//Grab the data and split it at ", "
		thisLoad := make([]byte, minInt64(loadChunk, size-loading))
		filein.ReadAt(thisLoad, loading)
		stringVals := strings.Split(lastVal+string(thisLoad), ", ")
		//Keep for the next run
		if loading+loadChunk < size {
			lastVal = stringVals[len(stringVals)-1]
		}
		//Go through the primes
		for index, val := range stringVals {
			//Break if we are at the end
			if index+2 > len(stringVals) && loading+loadChunk < size {
				break
			}
			//Parse the number and append it
			thisPrime, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				panic(err)
			}
			primes = append(primes, thisPrime)
		}
		//If we dont need to load anymore, then say we are continuing and break
		if primes[len(primes)-1] >= loadUpTo {
			isCont = true
			break
		}
	}
	//If we are continuing, we need to get the last prime in the file
	if isCont {
		//Load in the last number and parse it
		thisLoad := make([]byte, minInt64(size, 100))
		filein.ReadAt(thisLoad, size-int64(len(thisLoad)))
		stringVals := strings.Split(string(thisLoad), ", ")
		thisPrime, err := strconv.ParseUint(stringVals[len(stringVals)-1], 10, 64)
		if err != nil {
			panic(err)
		}
		//Maybee we are already done...
		if thisPrime >= finishAt {
			fmt.Println("You already have enough prime numbers lol...")
			os.Exit(0)
		}
		//Start at the next odd number
		current = thisPrime + 2
	} else {
		//Start at the next odd number
		current = primes[len(primes)-1] + 2
	}
	//If we are appending to the file, set up the globals for the appending
	if loadFile == saveFile {
		//We already have the size
		/*stats, err := os.Stat(loadFile)
		if err != nil {
			panic(err)
		}
		loc = stats.Size()*/
		//Start writing at the end of the file
		loc = size
		if !isCont {
			//Start writing from the 'next' prime
			writing = len(primes)
		}
		//YAY!!!! We are continuing to write!
		continueWrite = true
	}
	return
}

//Print the logo
func printLogo() {
	/*   ____       _                _____ _           _
		|  _ \ _ __(_)_ __ ___   ___|  ___(_)_ __   __| | ___ _ __
		| |_) | '__| | '_ ` _ \ / _ \ |_  | | '_ \ / _` |/ _ \ '__|
		|  __/| |  | | | | | | |  __/  _| | | | | | (_| |  __/ |
		|_|   |_|  |_|_| |_| |_|\___|_|   |_|_| |_|\__,_|\___|_|
	                                           By Jacob O'Toole    */

	fmt.Println(" ____       _                _____ _           _")
	fmt.Println("|  _ \\ _ __(_)_ __ ___   ___|  ___(_)_ __   __| | ___ _ __")
	fmt.Println("| |_) | '__| | '_ ` _ \\ / _ \\ |_  | | '_ \\ / _` |/ _ \\ '__|")
	fmt.Println("|  __/| |  | | | | | | |  __/  _| | | | | | (_| |  __/ |")
	fmt.Println("|_|   |_|  |_|_| |_| |_|\\___|_|   |_|_| |_|\\__,_|\\___|_|")
	fmt.Println("                                           By Jacob O'Toole\n")
}

func main() {
	//Load and save options
	doLoad := flag.Bool("l", false, "Load primes from text file")
	loadFile := flag.String("load", "primes.txt", "File to load primes from, -l flag must be present to load `file`s")
	flag.StringVar(&saveFile, "save", "primes.txt", "File to save primes to, if the `file` is the same as the load file, it will be appended to")
	flag.Int64Var(&loadChunk, "chunksize", 400, "Amount of `bytes` to load in for each chunk when loading primes in from a text file. Lower the value if the default uses to much RAM or is slower then expected. Only relevent when -l is used")
	//End options
	flag.Uint64Var(&finishAt, "max", math.MaxUint64, "Highest `number` to check. If -max is set, PrimeFinder will optimise itself for going up to this number. This can heavily reduce RAM usage and drasticaly improve start up speeds, so if you are struggling with RAM usage or want it to get up and running faster, set this to a value higher than you want/expect to get to and it will run as you want, but better. Defaults to "+strconv.FormatUint(math.MaxUint64, 10)+" - the miximum number PrimeFinder can check, but realistically higher then you will need.")
	flag.IntVar(&toFind, "end", math.MaxInt32, "Find the sepcified `amount` of prime numbers")
	//1 off options
	toCheck := flag.Uint64("check", 0, "Checks if the `number` given is prime. Then prints ONLY true/false depending on the output and exits. If the number given is 0, this flag will be ignored. Please note that this check is VERY inefficient as it does not rely on knowledge of previous primes as it does usaualy, meaning it checks if the number is a factor of 2 or ANY odd number up to the square root of the number - max value is "+strconv.FormatUint(math.MaxUint64, 10))
	getNearest := flag.Uint64("nearest", 0, "Finds the nearest prime `number`. Then prints ONLY this number. If the number given is 0, this flag will be ignored. Please note that this check is VERY inefficient as it does not rely on knowledge of previous primes as it does usaualy, meaning it checks if the number is a factor of 2 or ANY odd number up to the square root of the number - max value is "+strconv.FormatUint(math.MaxUint64, 10))
	nearestDirection := flag.String("direction", "both", "When finding the nearest prime (via the -nearest flag), this flag determines the search `direction`. Possible values are 'both', 'up' or 'down'. An invalid option will be ignored")
	//ANSI options
	var useANSI bool
	var niceLook bool
	if runtime.GOOS == "windows" {
		flag.BoolVar(&useANSI, "ansi", false, "Use ANSI to provide a nice UI - note that not all consoles support this, defaults to false on windows and true on linux")
		flag.BoolVar(&niceLook, "pretty", false, "Alias for -ansi (if either is true, ansi will be used)")
	} else {
		flag.BoolVar(&useANSI, "ansi", true, "Use ANSI to provide a nice UI - note that not all consoles support this, defaults to false on windows and true on linux")
		flag.BoolVar(&niceLook, "pretty", false, "Alias for -ansi (if either is true, ansi will be used)")
	}
	//Usage function
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "\nCongrats! You have accualy started to read the help page, that is the first step to becoming good at using the terminal...\nAnd as your reward, here are the flags:\n")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "Runtime info:")
		fmt.Fprintln(os.Stderr, "OS: "+runtime.GOOS)
		fmt.Fprintln(os.Stderr, "Architecture: "+runtime.GOARCH)
		fmt.Fprintln(os.Stderr, "Number of logical CPUs available to PrimeFinder: "+strconv.Itoa(runtime.NumCPU()))
		fmt.Fprintln(os.Stderr, "Go version: "+runtime.Version())
	}
	//Parse the flags
	flag.Parse()
	useANSI = useANSI || niceLook
	finishNumber := strconv.FormatUint(finishAt, 10)
	getting := strconv.FormatInt(int64(toFind), 10)
	//If we are checking somthing
	if *toCheck != 0 {
		if hardIsPrime(*toCheck) {
			fmt.Print("true")
		} else {
			fmt.Print("false")
		}
		return
	} else if *getNearest != 0 {
		//If we already are prime, no need to check
		if hardIsPrime(*getNearest) {
			fmt.Print(*getNearest)
			return
		}
		var current uint64 = 0
		for {
			//Check +current if we are going up, and/or -current if we are going down
			if (*nearestDirection == "both" || *nearestDirection == "up") && hardIsPrime(*getNearest+current) {
				fmt.Print(*getNearest + current)
				return
			} else if (*nearestDirection == "both" || *nearestDirection == "down") && hardIsPrime(*getNearest-current) {
				fmt.Print(*getNearest - current)
				return
			}
			current++
		}
	}
	//Set up the lines and the title
	if useANSI {
		fmt.Print("\u001b[2J\u001b[;H")
		printLogo()
		line1 = "\u001b[9;1H\u001b[2K"
		line2 = "\u001b[10;1H\u001b[2K"
		line3 = "\u001b[11;1H\u001b[2K"
	} else {
		printLogo()
		fmt.Println("!!! ANSI is disabled, enabling ANSI (via the -ansi switch) on a console that supports it makes the UI look nicer :)")
		line1 = "\r"
		line2 = "\r"
		line3 = "\r"
	}
	var amountLoaded uint64
	//Decide which log level to use
	var logLevel uint8
	//If we are loading, load, but if not, start with 2 as the only prime
	if *doLoad {
		amountLoaded = load(*loadFile, line1)
		//If we are finishing at a prime and dont have all the primes loaded, log level 2
		//But, if we are finishing at a number and DO have all the primes loaded, log level 1
		if isCont {
			logLevel = 2
		} else if finishAt != math.MaxUint64 {
			logLevel = 1
		}
	} else {
		primes = append(primes, 2)
	}
	//Run the garbidge collector before run to make sure everything is good.
	runtime.GC()
	//If log level is 2, then we dont have all the primes, so dont know how many we have loaded...
	if logLevel == 2 {
		//If they have supplied a -end option, we can't follow that :(
		if toFind != math.MaxInt32 {
			fmt.Println(line1 + "!!! Please note that the -end option has been ignored. This is to optimise for the -max value that has been set by the user. Increase the -max value or don't set it at all to not ignore -end.")
			if useANSI {
				line1 = "\u001b[10;1H\u001b[2K"
				line2 = "\u001b[11;1H\u001b[2K"
				line3 = "\u001b[12;1H\u001b[2K"
			}
		}
		//And we dont know how many primes we are starting with
		fmt.Println(line1 + "Started at " + strconv.FormatUint(current, 10) + " (loaded " + strconv.FormatUint(amountLoaded, 10) + " primes)...")
	} else {
		fmt.Println(line1 + "Started at " + strconv.FormatUint(current, 10) + " with " + strconv.Itoa(len(primes)) + " primes (loaded " + strconv.FormatUint(amountLoaded, 10) + ")...")
	}
	//We don't want to save on the first time round.
	saveYet := false
	//Forever - just like scratch
	for {
		//If the number we are checking is prime, append it to the correct list
		if isPrime(current) {
			if !isCont {
				primes = append(primes, current)
			} else {
				newPrimes = append(newPrimes, current)
			}
		}
		//Go up the the next odd number
		current += 2
		//Save every 100000000, but not on the first run
		if (current-1)%25e6 == 0 {
			if saveYet {
				if (current-1)%5e7 == 0 {
					//If we arn't using ANSI, then we will have to go down to the next line
					if !useANSI {
						fmt.Print("\n")
					}
					save()
					//GC just in case
					runtime.GC()
				}
			} else {
				saveYet = true
			}
		}
		//Log every 500000 numbers checked
		if (current-1)%5e5 == 0 {
			if logLevel == 0 {
				fmt.Print(line2 + strconv.FormatUint(uint64((float64(len(primes))/float64(toFind))*100), 10) + "% complete. Found " + strconv.Itoa(len(primes)) + " primes out of " + getting + ".")
			} else if logLevel == 1 {
				fmt.Print(line2 + strconv.FormatUint(primes[len(primes)-1]/finishAt, 10) + "%/" + strconv.FormatUint(uint64((float64(len(primes))/float64(toFind))*100), 10) + "% complete. Largest prime so far is " + strconv.FormatUint(primes[len(primes)-1], 10) + " out of the maximum " + finishNumber + ". Found " + strconv.Itoa(len(primes)) + " primes out of " + getting + ".")
			} else if logLevel == 3 {
				fmt.Print(line2 + "Largest prime so far is " + strconv.FormatUint(primes[len(primes)-1], 10) + " out of the maximum " + finishNumber + ".")
			} else {
				fmt.Print(line2 + strconv.FormatUint(uint64(float64(current)/float64(finishAt)*100), 10) + "% complete. Largest prime so far is " + strconv.FormatUint(newPrimes[len(newPrimes)-1], 10) + " (checked up to " + strconv.FormatUint(current, 10) + ") out of the maximum " + finishNumber + ".")
			}
		} else if !isCont {
			//If we arn't continuing, se if we have checked enough or have found enough primes to be finished
			if len(primes) >= toFind || current >= finishAt {
				break
			}
		} else if len(newPrimes) > 0 && current >= finishAt {
			//If we are continuing, and we have some primes, and we have checked over our finish value, we are done.
			break
		}
	}
	//Log that we are done and save.
	save()
	fmt.Println("\nDone.")
}
