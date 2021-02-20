*** imgorg

Over 20 years of taking photos and screenshots, getting new computers, backing up everything from past computers to newer computers, and never deleting anything has created the necessity for `imgorg` and [dups](https://gitpub.com/jasontconnell/dups).

So, there might be an image named `jason_drunk.jpg` that was copied 3 or 4 times into different computers and hence exist multiple times with the same byte stream and maybe even the same modified date.

`imgorg`'s job, if it so chooses to accept, is to find all instances of `jason_drunk.jpg` as they are strewn about in filesystem land (matching up by content hash), and copy them all to a folder structure with year/month/day/ following its modified time, and then some folders indicating where it found it. This can be a big set of nested folders but it can later be hand organized, knowing that `imgorg` did the [grunt work](https://www.jasontconnell.com/2018/08/26/grunt-work-principle).

~~As of this writing, it will only copy and leave the originals alone. Because it's still in dev and I haven't perfected the folder structure yet.~~

This can now delete files and folders with the -delete flag.

## Usage

```
imgorg -base c:\base\dir -sub "sub1,sub2" -dst c:\out\dir -roots "albums,pictures" -ignore _vti_cnf -workers 10
```

### All Flags

- base is the base folder to search.
- sub is for when you want to search subfolders instead. In my case, they are all stored on my Google Drive, but I want to search a few subfolders in there rather than searching the entire Google Drive folder.
- dst is the output folder root where the new structure will be created.
- roots is for when an image is contained within a folder called "albums" or "pictures" in this example, the output folder will be "albums" or "pictures" instead of the date structure. So that if you have all of your images for "Cancun Trip" in "albums", it will keep the structure `albums/Cancun Trip/`.
- map is a csv in the form of `folder=newfolder` which will map any files found in folder into newfolder. For example, `Phone Backup=Mobile Pics,iPhone Backup=Mobile Pics,Blackberry=Mobile Pics,Jason Phone=Mobile Pics`. So you can find all of your mobile pics in one place. Roots will take precedence but I'm not certain I'm cemented on this decision.
- ignore is folder names to ignore. Once it finds one it will skip all subdirectories of that as well. For example, -ignore _vti_cnf
- workers is the number of workers. Since no two workers try to work on the same file, it can have any number and be safe. Default is 20. It zips :)
- dryrun will run the read and just output changes it would make to the file system without doing it.
- delete will delete the src files after copying
- verbose will output all the things. Well, all the things I remembered to output when this flag is set.


### Live Example
Here's a live example in powershell of how I call `imgorg`.

```
go build

./imgorg `
    -base "C:\Users\jconnell\Google Drive" `
    -sub "G4,Genevieve,March 2013,Nexus,iPad Backup,Phone Backup,Old Backups,S2 backup" `
    -dst "C:\Users\jconnell\Google Drive\OrganizedImages" `
    -roots albums,pictures,jtccom,canada2002,connells `
    -map "Digicam=Mobile Pics,Nexus=Mobile Pics,Phone Backup=Mobile Pics,S2 backup=Mobile Pics,iPad Backup=Mobile Pics,bbpics=Mobile Pics,Camera Backup=Mobile Pics" `
    -ignore cutUps_v2 `
    -verbose `
    -delete `
> out.txt
```

And when you're done, you hopefully see something like this

`2021/02/20 01:26:02 finished. read: 14737534545 wrote: 14660556237 1m30.7191773s`

My Google Drive folder is syncing as I type this, and all of my images are much more organized!
