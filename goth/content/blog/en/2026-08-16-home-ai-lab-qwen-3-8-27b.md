---
title: "My Home AI Lab: Two GPUs, Proxmox, and a Local LLM in My Pocket"
description: "Why I built a dual-GPU AI server at home, how Proxmox, Ollama, and Qwen 3.8 27B fit together, and what it is like to run my own AI chat from a phone without cloud costs."
publishedAt: "2026-08-16"
slug: "home-ai-lab-qwen-3-8-27b"
draft: false
tags: ["AI", "home lab", "Proxmox", "Ollama", "hardware", "coding"]
---

The AI world has been moving at a remarkable pace lately. After getting tired of recurring cloud subscriptions and API bills, I decided to take matters into my own hands. I wanted to build a self-contained AI and development server at home: one that could run demanding language models locally, provide a remote desktop, and come with me wherever I went through my phone.

This is the story of how my weekend project progressed from assembling hardware to putting my own AI chat application into use.

### 1. Building the hardware: Why two RTX 5060 Ti cards?

I spent two evenings assembling and cabling my new PC. The centerpiece of the build is the GPU setup: I installed two Nvidia RTX 5060 Ti 16 GB cards.

When running large language models (LLMs), raw GPU compute is only part of the equation. The most important bottleneck is often VRAM, the graphics card's video memory. To keep inference fast, the model needs to fit in GPU memory rather than constantly moving data back and forth to system RAM.

One 16 GB graphics card quickly becomes restrictive once you move up to larger models in the 27 to 32 billion parameter range. Two 16 GB cards give me 32 GB of VRAM in total. That makes it possible to run larger models and broader context windows without spilling data into much slower system memory. Two separate cards also give me more flexibility when assigning resources to virtual machines.

### 2. Proxmox VE: A foundation for two operating systems

Once the hardware was ready, I installed Proxmox VE (Virtual Environment) as the foundation. Proxmox is a Type 1 hypervisor: it runs directly on the bare metal instead of on top of a separate host operating system. This lets me divide the machine's physical resources between isolated virtual machines (VMs).

I created two virtual machines in Proxmox:

- **Windows 11 (VM 200):** For general use, testing, and a few specialist applications.
- **Pop!_OS Linux (VM 100):** An Ubuntu-based distribution with a strong focus on Nvidia drivers. It is my primary development environment and AI workstation.

With PCI passthrough, I can assign the physical GPUs directly to the virtual machines. Windows and Pop!_OS can run at the same time, with each system getting its own hardware-accelerated graphics card.

### 3. Smooth remote access with Sunshine and Moonlight

I do not always want to sit in the same room as the server, so I wanted a convenient way to control it from my laptop. Traditional Windows RDP and Linux VNC often feel sluggish and choppy for active desktop work and coding.

The solution was the open-source combination of **Sunshine** and **Moonlight**:

- **Sunshine** runs as the host application on the server. It captures the desktop and encodes it into a video stream using Nvidia's hardware NVENC encoder, with very little CPU overhead.
- **Moonlight** is the client application on my laptop.

With this setup, the Pop!_OS desktop streams to my laptop over the home network at 60 to 120 fps with practically no noticeable latency. It feels just like sitting in front of the machine itself.

### 4. Getting the AI engine running: Ollama, Qwen 3.8 27B, and context windows

Then I got to the main event: running a local language model. I use Ollama to manage the AI environment. It is essentially Docker for language models: it downloads, configures, and runs open-source models directly from a lightweight terminal workflow.

I chose Alibaba's **Qwen 3.8 27B** as my main model. With 27 billion parameters, it offers remarkably strong reasoning and coding assistance for its size.

With 32 GB of VRAM available, I had enough room to create three variants of the model using Ollama `Modelfile` configurations, each with a different context window:

- **16k context:** Lightweight and extremely fast for short questions and small code snippets.
- **32k context:** The everyday sweet spot: enough room to keep dozens of pages of code or documentation in context.
- **64k context:** The heavy-duty option for analyzing large collections of files, using nearly all available VRAM.

### 5. Coding with Kilo Code and experiencing TTFT firsthand

I use VS Code as my development environment, together with the Kilo Code agent extension. It works in much the same way as other AI-assisted development tools, but instead of routing requests to a cloud service, it sends them across my local network to Ollama.

While coding, I noticed a characteristic of running a 27B model locally: TTFT, or *Time to First Token*. When I give the model a large source file, the GPUs take a few seconds to process the input and prepare the context in memory. Once that preparation is complete, however, the response and clean, usable code start appearing on screen surprisingly quickly.

### 6. A real-world test: Building my own chat app with two prompts

To test the system under realistic conditions, I gave my local Qwen 3.8 model a practical task: *"Build me a working web chat application that uses the Ollama API running on my server as its backend engine."*

The result genuinely surprised me. The model produced the backend, frontend, and necessary error handling in practice with just two consecutive prompts. There was no long round of manual bug fixing or tedious tweaking. I had working code ready to run.

### 7. The final piece: Tailscale puts the AI in my pocket

What is the point of a powerful home server if I can only use it from the couch? I solved secure remote access by installing Tailscale on my devices.

Tailscale is a straightforward, secure mesh VPN built on the WireGuard protocol. It connects my devices to the same protected virtual network without opening ports on my home router or exposing public IP addresses.

The complete setup now works like this: I enable Tailscale on my phone and open the server's internal IP address in a browser. My own chat application appears on the screen, ready to answer questions or help with coding from wherever I happen to be. The response is generated by the two-GPU server humming in my living room. My data does not pass through a third-party cloud service, and I do not pay a monthly subscription for it.

### Summary and ratings

This project was one of the most educational and rewarding things I have worked on in a long time. The hardware and software came together in a way that demonstrated, in practical terms, how capable local AI has become in 2026.

If I had to rate the results of the weekend, they would all get full marks:

- ⭐️⭐️⭐️⭐️⭐️ **Proxmox VE:** An exceptionally flexible foundation for virtual machines and GPU passthrough.
- ⭐️⭐️⭐️⭐️⭐️ **Qwen 3.8 27B:** Among the strongest open-source models I have tried. It codes accurately and handles large bodies of context well.
- ⭐️⭐️⭐️⭐️⭐️ **2x RTX 5060 Ti 16GB:** 32 GB of VRAM at this price point is a compelling sweet spot for anyone interested in building a home AI lab.

---

### Home server components

For reference, here is the complete parts list:

- **Graphics cards (2):** PNY GeForce RTX 5060 Ti 16GB GDDR7 (32 GB VRAM total)
- **Processor:** AMD Ryzen 7 7700
- **CPU cooler:** Thermalright Phantom Spirit 120 SE
- **Motherboard:** ASUS ProArt B850-CREATOR WIFI NEO
- **Memory (RAM):** Patriot 32GB (2x16GB) DDR5 6000MHz CL30
- **Storage (SSD):** Sandisk / WD Blue SN5100 1TB NVMe M.2 SSD (PCIe 4.0)
- **Power supply:** MSI MPG A1000G PCIE5 1000W Gold (ATX 3.1)
- **Case:** Lian Li LANCOOL 216 RGB (excellent airflow)
