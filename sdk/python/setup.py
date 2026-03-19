"""TokMan Python SDK Setup"""

from setuptools import setup, find_packages

setup(
    name="tokman",
    version="1.2.0",
    description="Token reduction with 14-layer research-based compression pipeline",
    author="GrayCodeAI",
    author_email="support@graycode.ai",
    url="https://github.com/GrayCodeAI/tokman",
    license="MIT",
    packages=find_packages(),
    python_requires=">=3.8",
    classifiers=[
        "Development Status :: 5 - Production/Stable",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
    ],
    keywords="token compression llm ai nlp",
)
