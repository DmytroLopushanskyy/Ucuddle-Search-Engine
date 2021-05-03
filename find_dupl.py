import json

with open("out.json", "r") as file:
	lines = file.readlines()
	links = [item for item in lines if ("\"link\":" in item)]
	doubled = [i for i in links if (links.count(i)>1)]
	for da in doubled:
		print(da, links.count(da))

with open("out.json", "r") as file:
	d = json.load(file)

for item in d:
	dd = [it for it in item["Hyperlinks"] if (item["Hyperlinks"].count(it) > 1)]
	for i in dd:
		print(i)

