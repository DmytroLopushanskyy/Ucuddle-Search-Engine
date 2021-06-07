import os
import sys
import subprocess

with open("for_py.txt", "r") as file:
    n = file.readlines()

con = 109

while len(n) > con:
    rite = n[:110]
    n = n[110:]
    with open('links.txt', 'w') as f:
        for item in rite:
            f.write("%s" % item)

    p1 = os.popen("go test -bench=. >>result.txt", "w")
    # p = subprocess.Popen(["go","test","-bench=.",">",">","result.txt"])
    # print "Happens while running"
    # p.communicate() #now wait plus that you can send commands to process
    
    print(len(n))