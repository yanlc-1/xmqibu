#!/usr/bin/env python3

import argparse
import json
import os
import sys
import urllib.error
import urllib.request


AUTHOR = "XMQibu Editorial Bot"


def build_article():
    return {
        "external_id": "editorial-github-weekly-top8-2026-05-27",
        "title": "GitHub 最近一周 Star 增长 Top 8：开发者正在追哪些 AI 与效率项目？",
        "summary": "基于 GitHub Trending 周榜整理 8 个最近一周涨星最快的项目，覆盖代码知识图谱、个人 AI 助手、Agent Memory、学术研究工作流和知识工作插件等方向。",
        "body": """这份榜单基于 GitHub Trending 周榜（统计窗口为最近一周）整理，时间点为 2026 年 5 月 27 日。我们更关心的不只是“谁涨得快”，还关心这些项目分别代表了哪些开发趋势：代码理解、Agent 基础设施、知识工作插件、个人 AI 助手，以及把 AI 变成实际生产力工具的各种尝试。

1. codegraph｜21,211 星/周
这个项目的定位非常清晰：先把代码库做成可查询的语义知识图谱，再让 Claude Code 一类的编程智能体直接基于图谱探索代码，而不是反复扫文件。对大仓库开发团队来说，它瞄准的是上下文成本和工具调用成本两个痛点。它这周冲得这么快，本质上反映了一个趋势：大家已经不满足于“模型会写代码”，而是在补 AI 编程的基础设施。

2. Understand-Anything｜19,191 星/周
这是一个把代码库或知识库转成交互式知识图谱的项目，强调“不是做一张好看的图，而是帮助人真正理解系统”。它支持多平台 AI 编程工具接入，还能做代码问答、diff 影响分析和新同学 onboarding。它爆发的原因很直观：开发者对“理解现有系统”这件事的需求，已经和“生成新代码”一样强。

3. AI Engineering from Scratch｜11,840 星/周
这是一个面向 AI 工程实践的大型开源课程，从线性代数一路到 autonomous agent swarms，强调“边学边做边交付”。它不是单点工具，而是一套系统化学习路径。最近热度高，说明市场对“AI 工程化教育”仍然极度饥渴，尤其是希望把模型、Agent、工具链、部署和实战放在一套课程里的人群。

4. OpenHuman｜8,542 星/周
OpenHuman 想做的是“个人 AI 超级智能体”，强调私有化、简单接入和日常生活中的持续陪伴。它不是企业工作流，而是站在个人用户视角，把 AI 当作长期使用的操作层。它的上涨说明一件事：除了 coding agent，面向个人工作与生活场景的 agent OS 也开始重新升温。

5. academic-research-skills｜8,422 星/周
这是一个偏学术研究工作流的插件/技能仓库，覆盖 research、write、review、revise、finalize 等环节，强调多 agent 协作、写作质量检查和引用完整性。它上涨很有代表性，因为科研、论文写作、系统综述这些高结构化知识工作，正在成为 Agent 最容易先落地的一批场景。

6. RuView｜5,986 星/周
RuView 做的是“无摄像头空间感知”：利用普通 WiFi 信号做空间智能、生命体征监测和 presence detection。虽然它看起来偏硬科技，但它走红说明 GitHub 热度并不只被纯软件 Agent 吞掉，AI 与传感、边缘设备、隐私友好的感知系统结合，仍然有很强吸引力。

7. agentmemory｜4,444 星/周
这是一个专门给 AI 编程智能体做持久化记忆的项目，强调本地 markdown 存储、语义检索和面向真实 benchmark 的 memory 设计。它火起来很正常，因为长任务、多轮修复、跨天持续开发，都会暴露“模型没有稳定记忆”的问题。谁先把 memory 做顺，谁就更接近真正可用的编程 Agent。

8. knowledge-work-plugins｜4,086 星/周
Anthropic 开源的这个仓库，把插件、连接器、命令和 skills 封装成按岗位分发的工作插件，面向 PM、销售、客服、法务、财务、数据等知识工作者。它的价值不在于某一个插件，而在于展示了一条路径：让 AI 不再只回答问题，而是带着组织流程、数据连接和岗位上下文去完成工作。

从这周榜单可以看出，GitHub 的“AI 热点”已经从单纯比拼模型效果，转向比拼三种能力：第一是理解复杂系统，第二是让 Agent 在真实工作流里持续运行，第三是把 AI 嵌进具体角色和行业场景。对内容站来说，这类榜单非常适合持续追踪，因为它天然能够提前暴露下一波开发者工具和 AI 应用的迁移方向。""",
        "author": AUTHOR,
        "source_links": [
            "https://github.com/trending?since=weekly",
            "https://github.com/colbymchenry/codegraph",
            "https://github.com/Lum1104/Understand-Anything",
            "https://github.com/rohitg00/ai-engineering-from-scratch",
            "https://github.com/tinyhumansai/openhuman",
            "https://github.com/Imbad0202/academic-research-skills",
            "https://github.com/ruvnet/RuView",
            "https://github.com/rohitg00/agentmemory",
            "https://github.com/anthropics/knowledge-work-plugins",
        ],
        "tags": ["github-top"],
        "published_at": "2026-05-27T10:00:00+08:00",
        "featured": True,
    }


def build_briefs():
    return [
        {
            "external_id": "brief-cursor-gartner-2026-05-22",
            "title": "Cursor 被 Gartner 列为 2026 企业 AI 编程智能体魔力象限领导者",
            "summary": "5 月 22 日，Cursor 表示其被 Gartner 评为 2026 年企业 AI 编程智能体领导者，并称已有超过 70% 的《财富》500 强企业在软件生命周期中部署和管理 Cursor 编程智能体。这说明 AI coding agent 正从个人工具走向组织级平台。",
            "external_url": "https://cursor.com/blog/cursor-leads-gartner-mq-2026",
            "importance": 5,
            "source_links": ["https://cursor.com/blog/cursor-leads-gartner-mq-2026"],
            "tags": ["overseas-ai"],
            "published_at": "2026-05-22T12:00:00+08:00",
        },
        {
            "external_id": "brief-anthropic-stainless-2026-05-18",
            "title": "Anthropic 收购 Stainless，补强 SDK 与 MCP 工具链",
            "summary": "Anthropic 于 5 月 18 日宣布收购 Stainless。官方强调，智能体能力不仅取决于模型本身，也取决于它能否高质量连接 API、SDK、CLI 和 MCP 服务器。这笔收购释放出一个明确信号：Agent 平台正在向“连接能力基础设施”继续加码。",
            "external_url": "https://www.anthropic.com/news/anthropic-acquires-stainless",
            "importance": 5,
            "source_links": ["https://www.anthropic.com/news/anthropic-acquires-stainless"],
            "tags": ["overseas-ai"],
            "published_at": "2026-05-18T12:00:00+08:00",
        },
        {
            "external_id": "brief-openai-voice-models-2026-05-07",
            "title": "OpenAI 推出新一代实时语音模型，覆盖推理、翻译与转写",
            "summary": "OpenAI 于 2026 年 5 月 7 日发布 GPT-Realtime-2、GPT-Realtime-Translate 和 GPT-Realtime-Whisper，面向实时语音交互、实时多语翻译和低延迟转写。语音接口开始从“能说话”升级到“能推理、能调用工具、能持续对话”。",
            "external_url": "https://openai.com/index/advancing-voice-intelligence-with-new-models-in-the-api/",
            "importance": 5,
            "source_links": ["https://openai.com/index/advancing-voice-intelligence-with-new-models-in-the-api/"],
            "tags": ["overseas-ai", "ai-products"],
            "published_at": "2026-05-07T12:00:00+08:00",
        },
        {
            "external_id": "brief-google-io-2026-ai-2026-05-20",
            "title": "Google I/O 2026 集中发布 Gemini 3.5、代理与开发工具更新",
            "summary": "Google 在 5 月 20 日汇总 I/O 2026 的 100 项发布，重点包括 Gemini 3.5 Flash、面向 agent-first 的开发平台，以及搜索、创作、购物和开发工具层面的 AI 更新。Google 正在把模型能力更紧密地打进产品和开发者入口。",
            "external_url": "https://blog.google/innovation-and-ai/technology/ai/google-io-2026-all-our-announcements/",
            "importance": 5,
            "source_links": ["https://blog.google/innovation-and-ai/technology/ai/google-io-2026-all-our-announcements/"],
            "tags": ["overseas-ai", "ai-products"],
            "published_at": "2026-05-20T12:00:00+08:00",
        },
        {
            "external_id": "brief-apple-intelligence-accessibility-2026-05-19",
            "title": "Apple 用 Apple Intelligence 强化无障碍功能，AI 更深嵌入终端体验",
            "summary": "Apple 于 2026 年 5 月 19 日预告一组由 Apple Intelligence 驱动的无障碍更新，包括更细粒度的视觉描述、自然语言导航和自动生成字幕。相比单独的聊天助手路径，苹果的打法更像是把 AI 逐层嵌进系统能力。",
            "external_url": "https://www.apple.com/newsroom/2026/05/apple-unveils-new-accessibility-features-and-updates-with-apple-intelligence/",
            "importance": 4,
            "source_links": ["https://www.apple.com/newsroom/2026/05/apple-unveils-new-accessibility-features-and-updates-with-apple-intelligence/"],
            "tags": ["overseas-ai", "ai-products"],
            "published_at": "2026-05-19T12:00:00+08:00",
        },
        {
            "external_id": "brief-meta-risk-review-2026-03-31",
            "title": "Meta 把 AI 引入产品风控评审，希望更早发现隐私与安全风险",
            "summary": "Meta 在 2026 年 3 月 31 日披露，其 Risk Review 体系已经把 AI 放入核心流程，用于提前识别隐私、安全和合规问题，并辅助审查文档预填与规则匹配。AI 在大厂内部的价值，正在从生成内容转向运营与治理效率。",
            "external_url": "https://about.fb.com/news/2026/03/how-ai-is-ushering-in-the-next-era-of-risk-review-at-meta/",
            "importance": 3,
            "source_links": ["https://about.fb.com/news/2026/03/how-ai-is-ushering-in-the-next-era-of-risk-review-at-meta/"],
            "tags": ["overseas-ai", "ai-products"],
            "published_at": "2026-03-31T12:00:00+08:00",
        },
        {
            "external_id": "brief-alibaba-ai-cloud-2026-03-19",
            "title": "阿里披露 AI + Cloud 进展：Qwen App 与云智能业务继续提速",
            "summary": "阿里巴巴 2026 年 3 月 19 日表示，其 AI 与云业务继续推进，Qwen App 在消费侧拉动采用，云智能集团收入增速加快。官方强调其全栈能力覆盖基础模型、云基础设施和自研芯片，目标是同时做企业 AI 与消费 AI。",
            "external_url": "https://www.alibabagroup.com/en-US/document-1971445322303406080",
            "importance": 4,
            "source_links": ["https://www.alibabagroup.com/en-US/document-1971445322303406080"],
            "tags": ["china-ai", "ai-products"],
            "published_at": "2026-03-19T12:00:00+08:00",
        },
        {
            "external_id": "brief-tencent-qclaw-global-2026-04-21",
            "title": "腾讯推出 QClaw 海外版，把个人 AI Agent 部署门槛压到 3 分钟",
            "summary": "腾讯于 2026 年 4 月 21 日开启 QClaw 国际版内测。官方称用户可在几分钟内完成安装，并把 AI Agent 接到 WhatsApp 或 Telegram 等通道。它代表的不是单一模型更新，而是“个人 Agent 产品化”进一步落地。",
            "external_url": "https://www.tencent.com/en-us/articles/2202318.html",
            "importance": 5,
            "source_links": ["https://www.tencent.com/en-us/articles/2202318.html"],
            "tags": ["china-ai", "ai-products"],
            "published_at": "2026-04-21T12:00:00+08:00",
        },
        {
            "external_id": "brief-tencent-hy3-preview-2026-04-24",
            "title": "腾讯发布并开源 Hy3 preview，强化推理、长上下文与 Agent 能力",
            "summary": "腾讯 2026 年 4 月 24 日推出并开源 Hy3 preview。官方把重点放在真实场景中的推理、长上下文、指令遵循、工具使用和成本效率，而不仅仅是 benchmark。这是国内大模型继续往“可用性”和“Agent 化”推进的典型信号。",
            "external_url": "https://www.tencent.com/en-us/articles/2202320.html",
            "importance": 4,
            "source_links": ["https://www.tencent.com/en-us/articles/2202320.html"],
            "tags": ["china-ai"],
            "published_at": "2026-04-24T12:00:00+08:00",
        },
        {
            "external_id": "brief-bytedance-seeduplex-2026-04-09",
            "title": "字节 Seed 发布全双工语音大模型 Seeduplex，并已在豆包 App 大规模上线",
            "summary": "字节 Seed 团队于 2026 年 4 月 9 日介绍 Seeduplex，全双工语音模型可实现“边听边说”，并已在豆包 App 全量落地。官方强调它在抗干扰、端点检测和自然度上都优于上一代半双工方案，说明语音 AI 正从实验室走向高频消费产品。",
            "external_url": "https://seed.bytedance.com/en/blog/introducing-seed-full-duplex-speech-llm-attentive-listening-robust-interference-suppression-enabling-more-natural-interaction",
            "importance": 5,
            "source_links": ["https://seed.bytedance.com/en/blog/introducing-seed-full-duplex-speech-llm-attentive-listening-robust-interference-suppression-enabling-more-natural-interaction"],
            "tags": ["china-ai", "ai-products"],
            "published_at": "2026-04-09T12:00:00+08:00",
        },
        {
            "external_id": "brief-baidu-ai-business-q1-2026-05-18",
            "title": "百度 Q1：AI 驱动业务收入过半，ERNIE 5.1、DuMate 与 GenFlow 继续推进",
            "summary": "百度 2026 年 5 月 18 日披露，一季度 Core AI-powered Business 收入首次超过其通用业务收入的一半。公告同时提到 ERNIE 5.1、日常生产力 Agent DuMate、Miaoda 3.0、Famou Agent 2.0 以及文库/网盘的 GenFlow 4.0，显示百度正同时押注模型、云和应用层。",
            "external_url": "https://ir.baidu.com/news-releases/news-release-details/baidu-announces-first-quarter-2026-results",
            "importance": 5,
            "source_links": ["https://ir.baidu.com/news-releases/news-release-details/baidu-announces-first-quarter-2026-results"],
            "tags": ["china-ai", "ai-products"],
            "published_at": "2026-05-18T12:00:00+08:00",
        },
        {
            "external_id": "brief-minimax-m27-2026-03-18",
            "title": "MiniMax 发布 M2.7，主打模型“自我进化”与复杂 Agent 任务",
            "summary": "MiniMax 在 2026 年 3 月 18 日发布 M2.7，核心叙事是让模型深度参与自身迭代，包括构建复杂 Agent Harness、更新自身 memory，并通过强化学习驱动进一步优化。它强调的是面向真实软件工程和生产力场景的长任务能力。",
            "external_url": "https://www.minimaxi.com/news/minimax-m27-zh",
            "importance": 4,
            "source_links": ["https://www.minimaxi.com/news/minimax-m27-zh"],
            "tags": ["china-ai"],
            "published_at": "2026-03-18T12:00:00+08:00",
        },
    ]


def post_json(base_url, token, path, payload, dry_run):
    if dry_run:
        print(f"[dry-run] POST {path}")
        print(json.dumps(payload, ensure_ascii=False, indent=2))
        return

    request = urllib.request.Request(
        base_url.rstrip("/") + path,
        data=json.dumps(payload, ensure_ascii=False).encode("utf-8"),
        headers={
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json",
        },
        method="POST",
    )
    with urllib.request.urlopen(request, timeout=30) as response:
        body = response.read().decode("utf-8", "replace")
        print(f"[{response.status}] {path} :: {payload['title']}")
        print(body)


def main():
    parser = argparse.ArgumentParser(description="Seed curated editorial content into the frontier site.")
    parser.add_argument("--base-url", required=True, help="Site base URL, such as https://xmqibu.com")
    parser.add_argument("--token", required=True, help="Ingest token")
    parser.add_argument("--dry-run", action="store_true", help="Print payloads instead of sending them")
    args = parser.parse_args()

    article = build_article()
    cover_image = os.environ.get("CURATED_GITHUB_COVER_IMAGE", "").strip()
    if cover_image:
        article["cover_image"] = cover_image
    briefs = build_briefs()

    try:
        post_json(args.base_url, args.token, "/api/ingest/articles", article, args.dry_run)
        for brief in briefs:
            post_json(args.base_url, args.token, "/api/ingest/briefs", brief, args.dry_run)
    except urllib.error.HTTPError as err:
        sys.stderr.write(f"HTTP {err.code}: {err.read().decode('utf-8', 'replace')}\n")
        raise SystemExit(1)
    except urllib.error.URLError as err:
        sys.stderr.write(f"request failed: {err}\n")
        raise SystemExit(1)


if __name__ == "__main__":
    main()
