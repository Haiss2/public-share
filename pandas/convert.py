import pandas as pd
import os
 
data_dir = "data/"
data_list = list(
    filter(lambda f: f.startswith("future") or f.startswith("spot"), 
    os.listdir(data_dir)))
data_list.sort()

def main():
    for data_file in data_list:
        if "multi" not in data_file:
            chunks = data_file.split('_')
            chunks[1] = 'multi'
            multi_file = '_'.join(chunks)
            statistic(data_file, multi_file)

def statistic(single, multi):
    print("Process {}, multiple symbols file corresponding {}".format(single, multi))
    s_df = pd.read_json(data_dir + single)
    s_df.rename(columns = {'us':'single_ts_micro'}, inplace = True)

    m_df = pd.read_json(data_dir + multi)
    m_df.rename(columns = {'us':'multiple_ts_micro'}, inplace = True)

    merged = s_df.merge(m_df)
    merged["time_diff_single_multi"] = merged["single_ts_micro"] - merged["multiple_ts_micro"] 
    merged["time_diff_multi_single"] = merged["multiple_ts_micro"] - merged["single_ts_micro"]

    print("single file event count:", len(s_df))
    print("merged event count:", len(merged))
    print(merged[["single_ts_micro", "multiple_ts_micro", "time_diff_single_multi","time_diff_multi_single"]].describe())
    print("")

main()
