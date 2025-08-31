import os
import shutil

def load_strings_from_file(file_path):
    """
    从文件中加载字符串列表。
    
    :param file_path: 文件路径
    :return: 字符串列表
    """
    try:
        with open(file_path, 'r', encoding='utf-8') as file:
            strings = [line.strip() for line in file if line.strip()]
        return strings
    except FileNotFoundError:
        print(f"错误：文件 '{file_path}' 未找到。")
    except Exception as e:
        print(f"发生错误：{e}")
    return []

def find_missing_items(strings, target_directory):
    """
    查找指定路径下不存在的目录名称。
    
    :param strings: 字符串列表
    :param target_directory: 指定路径
    :return: 不在路径下的字符串列表
    """
    try:
        # 获取指定路径下的所有目录名称
        existing_dirs = set(os.listdir(target_directory))
        
        # 筛选出不在路径下的字符串
        missing_items = [item for item in strings if item not in existing_dirs]
        return missing_items
    
    except FileNotFoundError:
        print(f"错误：路径 '{target_directory}' 未找到。")
    except Exception as e:
        print(f"发生错误：{e}")
    return []

def copy_missing_directories(missing_items, source_directory, target_directory):
    """
    从源目录复制缺失的目录到目标目录。
    
    :param missing_items: 缺失的目录名称列表
    :param source_directory: 源目录路径
    :param target_directory: 目标目录路径
    """
    try:
        for item in missing_items:
            source_path = os.path.join(source_directory, item)
            target_path = os.path.join(target_directory, item)
            
            # 检查源目录中是否存在该目录
            if os.path.isdir(source_path):
                print(f"正在复制目录：'{item}' 从 '{source_directory}' 到 '{target_directory}'")
                shutil.copytree(source_path, target_path)
            else:
                print(f"- {item}")
    except Exception as e:
        print(f"复制过程中发生错误：{e}")

# 示例用法
if __name__ == "__main__":
    # 定义输入文件和路径
    input_file = "output_strings.txt"  # 前一步生成的字符串文件
    target_directory = "D://Code//Zbp//sucai//"  # 目标路径
    source_directory = "C://Users//19853//Desktop//materials//materials"  # 源路径
    
    # 加载字符串列表
    strings = load_strings_from_file(input_file)
    if not strings:
        print("未加载到任何字符串。")
        exit(1)
    
    # 查找不在路径下的字符串
    missing_items = find_missing_items(strings, target_directory)
    
    # 如果存在缺失的目录，则进行复制
    if missing_items:
        print("以下目录不存在，开始复制：")
        print(", ".join(missing_items))
        copy_missing_directories(missing_items, source_directory, target_directory)
    else:
        print("所有目录都已存在，无需复制。")