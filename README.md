### DiscordCacheExtractor

### Summary
A small console based tool to detect and convert files, saved in the discord cache.

### What can it do?
The program gets all files from the given discord cache folder, detects their filetype and copies them into a folder of your choice. 
> (Standard is the folder, where the application was started from.)

### Parameters
##### -src [Path]  

Enter the cache path of discord. This is the only neccessary parameter, the "-src" can be left out, if this path is the only parameter.

------------

##### -dst [Path] - Standard: Path of the execution file.

Enter the path, where the converted pictures should be saved.

------------

##### -cs [Num] - Standard: 10

To guarantee a good execution time, the work of copying is split up to several subprocesses. This number defines, how many files a subprocess handels at maximum.
- The lower, the less processes must be started, may take longer due to less work at the same time.
- The higher, the more processes must be started which also costs time. 

> You need to find a good average for your goals. Nevertheless, most of the time the program has to handle only a few hundred pictures, so one second more or less should not ruin your day. ;D

------------

##### -tc [Num] - Standard: very much

Enter the maximum number of subprocesses (explained above) that should run at the same time.

------------

##### -k - Standard: False

Marks that you also want to copy files with undetected filetype.


------------

##### -s [/ or \\]
Set the path seperator, used by your system.  
- Windows: \\\\
- Linux: /

------------

###### Last updated: 2021-02-11 14:04:08 Thursday
