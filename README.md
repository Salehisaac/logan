# Logan 

![Go version](https://img.shields.io/badge/go-v1.19.7-blue)

## Overview

Welcome to the Logan !, your key to efficient log management and request tracking. This tool is designed to simplify the process of monitoring and analyzing logs for specific users by organizing them based on trace IDs. Whether you're tracing request lifecycles or analyzing historical data, Logan offers a straightforward solution.

In log management, keeping track of requests and their corresponding logs can be complex. This tool helps you navigate through this complexity by providing a clear and organized view of log entries. It allows you to extract logs starting from a specific timestamp or in real-time, ensuring you capture every detail of the request lifecycle.

Logan operates in two modes:

Static Mode: Extracts logs from a specified start time up to the present. This mode is useful for historical analysis and data extraction.
Real-Time Mode: Activated with the -s flag, this mode processes and captures logs as they are generated, providing immediate insights into ongoing activities.
By utilizing Logan, you gain a deeper understanding of how requests are processed through your system. You'll be able to monitor request flows, troubleshoot issues, and ensure smooth operation by having a comprehensive view of log data.

## Installation

Follow these steps to set up and use the Log Reader tool:

1. **Clone the Repository**

   Start by cloning the Log Reader repository to your local machine:
   ```bash
   git clone https://github.com/Salehisaac/logan.git

2. **Navigate to the Project Directory**

    Change into the project directory:
     ```bash
     cd logan

3. **Set Up Environment Variables**

    Create a .env file in the project root directory. This file should contain your configuration settings. Use the provided .env.example as a template:
     ```bash
     cp .env.example .env
    ```
    Open .env and configure it according to your environment. Typical entries might include settings for log file paths, time zones, or other relevant configurations.

4. **Build the Project**

    Compile the project using the provided Makefile:
     ```bash
     make build

5. **Run the Log Reader**

    *Static Read:* To read logs from a specific start time until now, use the following command. Replace hr:min:sec with your desired start time:
    ```bash
    ./logan  --time hr:min:sec
    ```
  
    *Real-Time Read:* To capture logs as they are generated, use the -s flag:
    ```bash
    ./logan -s
    ```
