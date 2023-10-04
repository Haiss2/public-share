import pandas as pd
import os
import matplotlib.pyplot as plt
import time

list_data_20 = list(os.listdir("data 20"))
list_data_200 = list(os.listdir("data 200"))


def analyze(filename):
    print("Filename:", filename)
    check_20 = filename in list_data_20
    check_200 = filename in list_data_200
    print("Check file in data_20", check_20)
    print("Check file in data_200", check_200)
    if not check_20 or not check_200:
        return
    
    data20 = pd.read_json("data 20/" + filename)
    data200 = pd.read_json("data 200/" + filename)
    print("file in data_20 length", len(data20))
    print("file in data_200 length", len(data200))
    
    data20["actual_time_diff"] = (data200["actual_time"] - data20["actual_time"])

    print(data20[["actual_time_diff"]].describe())
    print("\n")

    data20["actual_time_diff"] .plot()
    plt.savefig("img/{}.png".format(filename))
    plt.show()


for filename in list_data_20:
    analyze(filename)
    break

