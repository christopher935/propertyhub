import pytest
import asyncio
import tempfile
import shutil
from pathlib import Path
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from bash_server import FileTool, BashTool, _BashSession, WORKSPACE_DIR

@pytest.fixture(scope="function")
def event_loop():
    """Create event loop for async tests."""
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    yield loop
    try:
        loop.run_until_complete(asyncio.sleep(0))
    except Exception:
        pass
    finally:
        loop.close()

@pytest.fixture
def temp_workspace(tmp_path):
    """Create a temporary workspace directory for testing."""
    workspace = tmp_path / "workspace"
    workspace.mkdir()
    yield workspace

@pytest.fixture
def file_tool(temp_workspace):
    """Create a FileTool instance with temp workspace."""
    return FileTool(base_path=temp_workspace)

@pytest.fixture
async def bash_tool():
    """Create a BashTool instance for testing."""
    tool = BashTool()
    yield tool
    for session in tool._sessions.values():
        session.stop()

@pytest.fixture
def sample_file(temp_workspace):
    """Create a sample file for testing."""
    file_path = temp_workspace / "sample.txt"
    file_path.write_text("line 1\nline 2\nline 3\n")
    return file_path

@pytest.fixture
def sample_dir(temp_workspace):
    """Create a sample directory structure for testing."""
    subdir = temp_workspace / "subdir"
    subdir.mkdir()
    (subdir / "file1.txt").write_text("content 1")
    (subdir / "file2.txt").write_text("content 2")
    return subdir
