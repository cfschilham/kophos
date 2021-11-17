import statistics
import numpy

def sum(list):
    sum = 0
    for i in list:
        sum += i
    return sum

def average(list):
    return sum(list) / len(list)

lines = open("/home/cfschilham/dullhash.csv", "r").readlines()
del lines[0]

hashes = []
datas = []

for line in lines:
    data, hash = line.split(",")
    hashes.append(int(hash))
    datas.append(int(data))

hAvg = average(hashes)
dAvg = average(datas)
hDev = statistics.stdev(hashes)
dDev = statistics.stdev(datas)

print(f"Hashes average: {hAvg}")
print(f"Input data average: {dAvg}")
print(f"Standard deviation hashes: {hDev}")
print(f"Standard deviation input data: {dDev}")

eX = []
eY = []
eXY = []

for hash in hashes:
    eX.append(hash + hAvg)
for data in datas:
    eY.append(data + dAvg)

i = 0
for x in eX:
    eXY.append(x * eY[i])
    i += 1

# c = numpy.cov(datas, hashes)
cc = numpy.corrcoef(numpy.asarray(datas).astype(float), numpy.asarray(hashes).astype(float))
# print(f"Covariance: {c}")
print(f"Correlation coefficient: {cc}")