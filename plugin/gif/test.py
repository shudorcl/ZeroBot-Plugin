import re

def extract_strings_from_file(file_path):
    """
    从指定的代码文件中提取 dlrange 函数调用的第一个字符串参数。
    
    :param file_path: 代码文件路径
    :return: 提取的字符串列表
    """
    # 正则表达式匹配 dlrange 函数调用的第一个字符串参数
    pattern = r'dlrange\s*\(\s*"([^"]+)"'
    strings = []

    try:
        # 打开并读取文件内容
        with open(file_path, 'r', encoding='utf-8') as file:
            content = file.read()
        
        # 使用正则表达式查找所有匹配的字符串
        matches = re.findall(pattern, content)
        strings.extend(matches)
    
    except FileNotFoundError:
        print(f"错误：文件 '{file_path}' 未找到。")
    except Exception as e:
        print(f"发生错误：{e}")
    
    return strings

def write_strings_to_file(strings, output_file):
    """
    将字符串列表写入到指定的输出文件中。
    
    :param strings: 字符串列表
    :param output_file: 输出文件路径
    """
    try:
        with open(output_file, 'w', encoding='utf-8') as file:
            for string in strings:
                file.write(f"{string}\n")
        print(f"字符串已成功写入到文件 '{output_file}' 中。")
    except Exception as e:
        print(f"写入文件时发生错误：{e}")

# 示例用法
if __name__ == "__main__":
    # 定义输入文件和输出文件
    input_files = ["png.go", "gif.go"]
    output_file = "output_strings.txt"
    
    # 汇总所有文件中的字符串
    all_strings = []
    for file in input_files:
        print(f"正在处理文件：{file}")
        strings = extract_strings_from_file(file)
        all_strings.extend(strings)
    
    # 去重（如果需要）并写入到输出文件
    unique_strings = list(set(all_strings))  # 如果需要去重
    write_strings_to_file(unique_strings, output_file)