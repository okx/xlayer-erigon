import re
import matplotlib.pyplot as plt

def parse_size_to_mb(size_str):
    """
    Convert size strings like '129MiB', '35.18GiB' to MB (float).
    """
    size_str = size_str.strip()
    match = re.match(r'([\d\.]+)([MG]i?B)', size_str, re.IGNORECASE)
    if not match:
        return None
    
    value = float(match.group(1))
    unit = match.group(2).upper()
    
    if unit == 'MIB':
        return value
    elif unit == 'GIB':
        return value * 1024
    elif unit == 'MB':
        return value
    elif unit == 'GB':
        return value * 1000
    return None

def parse_docker_stats_line(line):
    """
    Parse Docker stats line into a dictionary.
    """
    columns = line.split()
    cpu_str = columns[2].replace('%', '')
    cpu_float = float(cpu_str)

    mem_usage_str = columns[3]
    mem_perc = float(columns[6].replace('%', ''))

    mem_usage_mb = parse_size_to_mb(mem_usage_str)

    return {
        'cpu': cpu_float,
        'mem_usage_mb': mem_usage_mb if mem_usage_mb else 0.0,
        'mem_perc': mem_perc,
    }

def main():
    stats_data = []

    with open('seq_stats.txt', 'r', encoding='utf-8') as f:
        lines = f.readlines()

    for line in lines:
        line = line.strip()
        if not line:
            continue
        
        if 'xlayer-seq' in line and 'CONTAINER ID' not in line:
            parsed = parse_docker_stats_line(line)
            stats_data.append(parsed)

    # Extracting data for plotting
    indices = list(range(len(stats_data)))
    cpu_usages = [item['cpu'] for item in stats_data]
    mem_usages = [item['mem_usage_mb'] for item in stats_data]
    mem_percs = [item['mem_perc'] for item in stats_data]

    # Find max and min values for memory percentage, CPU, and memory usage
    max_mem_perc, min_mem_perc = max(mem_percs), min(mem_percs)
    max_cpu, min_cpu = max(cpu_usages), min(cpu_usages)
    max_mem_usage, min_mem_usage = max(mem_usages), min(mem_usages)

    max_mem_perc_idx, min_mem_perc_idx = mem_percs.index(max_mem_perc), mem_percs.index(min_mem_perc)
    max_cpu_idx, min_cpu_idx = cpu_usages.index(max_cpu), cpu_usages.index(min_cpu)
    max_mem_usage_idx, min_mem_usage_idx = mem_usages.index(max_mem_usage), mem_usages.index(min_mem_usage)

    # Plotting all metrics in a single chart
    fig, ax1 = plt.subplots(figsize=(14, 7))

    # CPU and Memory Usage Percentage (left y-axis)
    ax1.plot(indices, cpu_usages, marker='o', label='CPU Usage (%)', color='blue')
    ax1.plot(indices, mem_percs, marker='s', label='Memory Usage (%)', color='green')
    ax1.set_xlabel('Index')
    ax1.set_ylabel('Usage Percentage (%)')
    ax1.tick_params(axis='y')
    ax1.legend(loc='upper left')

    # Annotate max and min values
    ax1.annotate(f'Max CPU: {max_cpu}%', xy=(max_cpu_idx, max_cpu), xytext=(max_cpu_idx, max_cpu + 2),
                 arrowprops=dict(facecolor='blue', arrowstyle='->'), fontsize=10, color='blue')
    ax1.annotate(f'Min CPU: {min_cpu}%', xy=(min_cpu_idx, min_cpu), xytext=(min_cpu_idx, min_cpu - 2),
                 arrowprops=dict(facecolor='blue', arrowstyle='->'), fontsize=10, color='blue')

    ax1.annotate(f'Max Mem %: {max_mem_perc}%', xy=(max_mem_perc_idx, max_mem_perc), xytext=(max_mem_perc_idx, max_mem_perc + 2),
                 arrowprops=dict(facecolor='green', arrowstyle='->'), fontsize=10, color='green')
    ax1.annotate(f'Min Mem %: {min_mem_perc}%', xy=(min_mem_perc_idx, min_mem_perc), xytext=(min_mem_perc_idx, min_mem_perc - 2),
                 arrowprops=dict(facecolor='green', arrowstyle='->'), fontsize=10, color='green')

    # Memory Usage in MB (right y-axis)
    ax2 = ax1.twinx()
    ax2.plot(indices, mem_usages, marker='^', label='Memory Usage (MB)', color='orange')
    ax2.set_ylabel('Memory Usage (MB)')
    ax2.tick_params(axis='y')
    ax2.legend(loc='upper right')

    # Annotate max and min memory usage
    ax2.annotate(f'Max Mem: {max_mem_usage:.2f} MB', xy=(max_mem_usage_idx, max_mem_usage), 
                 xytext=(max_mem_usage_idx, max_mem_usage + 50),
                 arrowprops=dict(facecolor='orange', arrowstyle='->'), fontsize=10, color='orange')
    ax2.annotate(f'Min Mem: {min_mem_usage:.2f} MB', xy=(min_mem_usage_idx, min_mem_usage), 
                 xytext=(min_mem_usage_idx, min_mem_usage - 50),
                 arrowprops=dict(facecolor='orange', arrowstyle='->'), fontsize=10, color='orange')

    # Title and layout
    plt.title('CPU and Memory Usage Over Time (Max/Min Highlighted)')
    fig.tight_layout()
    plt.show()

if __name__ == '__main__':
    main()