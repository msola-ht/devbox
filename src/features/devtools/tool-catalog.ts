import type { Cancellable, Platform } from "mygo-runtime";

import { DevToolsService, type InstallTool, type ToolDetection } from "@/mygo";

interface ToolDefinition {
  id: InstallTool;
  name: string;
  description: string;
  detect: Cancellable<[boolean], ToolDetection[]>;
  platforms?: readonly Platform[];
}

export interface ToolGroupDefinition {
  id: string;
  title: string;
  tools: readonly ToolDefinition[];
}

const developmentToolGroups: readonly ToolGroupDefinition[] = [
  {
    id: "tools-system-packages",
    title: "系统包管理器",
    tools: [
      {
        id: "homebrew",
        name: "Homebrew",
        description: "macOS 系统包管理器",
        detect: DevToolsService.detectHomebrew,
        platforms: ["darwin"],
      },
      {
        id: "winget",
        name: "WinGet",
        description: "Windows App Installer 提供的包管理器",
        detect: DevToolsService.detectWinGet,
        platforms: ["win32"],
      },
      {
        id: "chocolatey",
        name: "Chocolatey",
        description: "Windows 软件包管理器",
        detect: DevToolsService.detectChocolatey,
        platforms: ["win32"],
      },
    ],
  },
  {
    id: "tools-general",
    title: "基础工具",
    tools: [
      { id: "git", name: "Git", description: "版本控制工具", detect: DevToolsService.detectGit },
      { id: "gh", name: "gh", description: "GitHub 命令行工具", detect: DevToolsService.detectGH },
      { id: "jq", name: "jq", description: "JSON 处理工具", detect: DevToolsService.detectJQ },
      { id: "ripgrep", name: "rg", description: "文本搜索工具", detect: DevToolsService.detectRG },
    ],
  },
  {
    id: "tools-python",
    title: "Python",
    tools: [
      { id: "uv", name: "uv", description: "Python 包与项目管理工具", detect: DevToolsService.detectUV },
      { id: "python", name: "Python", description: "Python 3 解释器", detect: DevToolsService.detectPython },
    ],
  },
  {
    id: "tools-javascript",
    title: "JavaScript",
    tools: [
      { id: "node", name: "Node.js", description: "JavaScript 运行时", detect: DevToolsService.detectNode },
      { id: "npm", name: "npm", description: "JavaScript 包管理器", detect: DevToolsService.detectNPM },
      { id: "pnpm", name: "pnpm", description: "JavaScript 包管理器", detect: DevToolsService.detectPNPM },
    ],
  },
  {
    id: "tools-shells",
    title: "Shell 与脚本",
    tools: [
      {
        id: "zsh",
        name: "Zsh",
        description: "命令行 Shell 与脚本环境",
        detect: DevToolsService.detectZsh,
        platforms: ["darwin", "linux"],
      },
      {
        id: "powershell",
        name: "PowerShell",
        description: "命令行 Shell 与脚本环境",
        detect: DevToolsService.detectPowerShell,
        platforms: ["win32"],
      },
    ],
  },
];

export function getDevelopmentToolGroups(platform: Platform | undefined) {
  return developmentToolGroups
    .map((group) => ({
      ...group,
      tools: group.tools.filter(({ platforms }) =>
        !platforms || (platform !== undefined && platforms.includes(platform))),
    }))
    .filter(({ tools }) => tools.length > 0);
}
