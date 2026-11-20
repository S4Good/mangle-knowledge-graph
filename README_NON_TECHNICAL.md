# Infrastructure Dependency Analysis — Explained Simply

## The Problem

Imagine your company uses several servers, applications, and databases that work together.

If one system stops working, you may not know which other applications or business processes could be affected.

This makes IT maintenance, troubleshooting, and risk planning more difficult.

## My Solution

I developed a tool called **Mangle Knowledge Graph** that analyzes known connections between IT systems.

It helps answer three important questions:

- **What does this system depend on?**
- **What other systems could be affected if it fails?**
- **Why are these systems connected?**

The tool uses logical rules to identify relationships, including indirect dependencies that may not be immediately obvious.

## Business Benefits

- **Better visibility:** Understand how IT systems depend on one another.
- **Risk awareness:** Identify systems whose failure could affect multiple applications.
- **Maintenance planning:** Understand potential consequences before making changes.
- **Clear explanations:** See why a particular system may be affected.

## A Practical Example

Suppose your company's customer portal depends on an application server, which in turn depends on a database.

If the database becomes unavailable, the tool can identify both the application server and customer portal as potentially affected.

This helps your IT team investigate risks and plan appropriate safeguards.

## Important Limitations

The tool analyzes infrastructure information and dependencies that have been supplied and validated. It does not automatically discover every connection, continuously monitor system health, or prevent outages.

**The goal is to help organizations understand their IT infrastructure and make better-informed decisions.**

---

[View the technical documentation](README.md)