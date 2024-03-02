# exampleslottedpage

File Structure: 
File Consists of FileMetadata at the start of the file and then individual pages , each page is of 4096 Bytes
![image](https://github.com/Vishwanath-V/exampleslottedpage/assets/53922593/a320f353-1eca-4186-8580-4c6f97aae588)


Page Struture: 
Page consists of Page Header Metadata at the start of the page and then data cell key offsets (called as slots) and then free space and then data cells for each slot starting from backside of the page
![image](https://github.com/Vishwanath-V/exampleslottedpage/assets/53922593/765165ef-d39a-4b2e-beeb-85c07a73500c)
