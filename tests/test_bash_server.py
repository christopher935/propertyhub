import pytest
import asyncio
from pathlib import Path
import os
import re

from bash_server import (
    FileTool, BashTool, _BashSession, 
    ToolResult, CLIResult, ToolError,
    validate_command_length
)


class TestFileTool:
    """Unit tests for FileTool class."""

    @pytest.mark.asyncio
    async def test_read_text_file(self, file_tool, sample_file, temp_workspace):
        """Test reading a text file."""
        result = await file_tool.read(str(sample_file.relative_to(temp_workspace)))
        assert result.output is not None
        assert "line 1" in result.output
        assert result.error is None

    @pytest.mark.asyncio
    async def test_read_with_line_numbers(self, file_tool, sample_file, temp_workspace):
        """Test reading with line numbers enabled."""
        result = await file_tool.read(
            str(sample_file.relative_to(temp_workspace)), 
            line_numbers=True
        )
        assert "1\t" in result.output or "1	" in result.output

    @pytest.mark.asyncio
    async def test_read_nonexistent_file(self, file_tool):
        """Test reading a file that doesn't exist."""
        result = await file_tool(command="read", path="nonexistent.txt")
        assert result.error is not None
        assert "not a file" in result.error.lower() or "invalid path" in result.error.lower()

    @pytest.mark.asyncio
    async def test_write_new_file(self, file_tool, temp_workspace):
        """Test writing a new file."""
        result = await file_tool.write("newfile.txt", "test content")
        assert "written" in result.output.lower()
        assert (temp_workspace / "newfile.txt").read_text() == "test content"

    @pytest.mark.asyncio
    async def test_write_overwrites_existing(self, file_tool, sample_file, temp_workspace):
        """Test that write overwrites existing content."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        await file_tool.write(rel_path, "new content")
        assert sample_file.read_text() == "new content"

    @pytest.mark.asyncio
    async def test_path_traversal_blocked(self, file_tool):
        """Test that path traversal attempts are blocked."""
        result = await file_tool(command="read", path="../../../etc/passwd")
        assert result.error is not None
        assert "outside" in result.error.lower() or "invalid" in result.error.lower()

    @pytest.mark.asyncio
    async def test_absolute_path_outside_workspace(self, file_tool):
        """Test that absolute paths outside workspace are blocked."""
        result = await file_tool(command="read", path="/etc/passwd")
        assert result.error is not None

    @pytest.mark.asyncio
    async def test_list_directory(self, file_tool, sample_dir, temp_workspace):
        """Test listing directory contents."""
        rel_path = str(sample_dir.relative_to(temp_workspace))
        result = await file_tool.list_dir(rel_path)
        assert "file1.txt" in result.output
        assert "file2.txt" in result.output

    @pytest.mark.asyncio
    async def test_mkdir_creates_nested(self, file_tool, temp_workspace):
        """Test that mkdir creates nested directories."""
        result = await file_tool.mkdir("a/b/c")
        assert result.error is None
        assert (temp_workspace / "a" / "b" / "c").is_dir()

    @pytest.mark.asyncio
    async def test_grep_finds_pattern(self, file_tool, sample_file, temp_workspace):
        """Test grep finding a pattern."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool.grep(pattern="line 2", path=rel_path)
        assert "line 2" in result.output

    @pytest.mark.asyncio
    async def test_grep_case_insensitive(self, file_tool, sample_file, temp_workspace):
        """Test case-insensitive grep."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool.grep(
            pattern="LINE", 
            path=rel_path, 
            case_sensitive=False
        )
        assert result.output is not None
        assert "line" in result.output.lower()

    @pytest.mark.asyncio
    async def test_grep_no_match(self, file_tool, sample_file, temp_workspace):
        """Test grep with no matches."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool.grep(pattern="nonexistent", path=rel_path)
        assert "no matches" in result.output.lower()

    @pytest.mark.asyncio
    async def test_replace_single(self, file_tool, sample_file, temp_workspace):
        """Test replacing a single occurrence."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool.replace(rel_path, "line 1", "LINE ONE")
        assert result.error is None
        content = sample_file.read_text()
        assert "LINE ONE" in content

    @pytest.mark.asyncio
    async def test_replace_all_occurrences(self, file_tool, temp_workspace):
        """Test replacing all occurrences."""
        test_file = temp_workspace / "test.txt"
        test_file.write_text("foo bar foo baz foo")
        result = await file_tool.replace("test.txt", "foo", "X", all_occurrences=True)
        assert result.error is None
        assert test_file.read_text() == "X bar X baz X"

    @pytest.mark.asyncio
    async def test_undo_restores_previous(self, file_tool, sample_file, temp_workspace):
        """Test that undo restores previous content."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        original = sample_file.read_text()
        await file_tool.replace(rel_path, "line 1", "CHANGED")
        result = await file_tool.undo(rel_path)
        assert result.error is None
        assert sample_file.read_text() == original

    @pytest.mark.asyncio
    async def test_append_to_file(self, file_tool, sample_file, temp_workspace):
        """Test appending content to a file."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        original = sample_file.read_text()
        result = await file_tool.append(rel_path, "line 4\n")
        assert result.error is None
        assert sample_file.read_text() == original + "line 4\n"

    @pytest.mark.asyncio
    async def test_delete_file(self, file_tool, sample_file, temp_workspace):
        """Test deleting a file."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool.delete(rel_path)
        assert result.error is None
        assert not sample_file.exists()

    @pytest.mark.asyncio
    async def test_exists_returns_true(self, file_tool, sample_file, temp_workspace):
        """Test exists returns True for existing file."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool.exists(rel_path)
        assert result.output == "True"

    @pytest.mark.asyncio
    async def test_exists_returns_false(self, file_tool):
        """Test exists returns False for nonexistent file."""
        result = await file_tool.exists("nonexistent.txt")
        assert result.output == "False"

    @pytest.mark.asyncio
    async def test_move_file(self, file_tool, sample_file, temp_workspace):
        """Test moving a file."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool.move(rel_path, "moved.txt")
        assert result.error is None
        assert not sample_file.exists()
        assert (temp_workspace / "moved.txt").exists()

    @pytest.mark.asyncio
    async def test_copy_file(self, file_tool, sample_file, temp_workspace):
        """Test copying a file."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool.copy(rel_path, "copied.txt")
        assert result.error is None
        assert sample_file.exists()
        assert (temp_workspace / "copied.txt").exists()
        assert sample_file.read_text() == (temp_workspace / "copied.txt").read_text()

    @pytest.mark.asyncio
    async def test_view_file(self, file_tool, sample_file, temp_workspace):
        """Test viewing a file."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool.view(rel_path)
        assert result.error is None
        assert "line 1" in result.output

    @pytest.mark.asyncio
    async def test_view_with_range(self, file_tool, sample_file, temp_workspace):
        """Test viewing a file with line range."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool.view(rel_path, view_range=[1, 2])
        assert result.error is None
        assert "line 1" in result.output
        assert "line 2" in result.output
        assert "line 3" not in result.output

    @pytest.mark.asyncio
    async def test_create_file(self, file_tool, temp_workspace):
        """Test creating a new file."""
        result = await file_tool.create("newfile.txt", "new content")
        assert result.error is None
        assert (temp_workspace / "newfile.txt").read_text() == "new content"

    @pytest.mark.asyncio
    async def test_create_file_already_exists(self, file_tool, sample_file, temp_workspace):
        """Test that create fails if file already exists."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool(command="create", path=rel_path, content="content")
        assert result.error is not None
        assert "already exists" in result.error.lower()

    @pytest.mark.asyncio
    async def test_insert_line(self, file_tool, sample_file, temp_workspace):
        """Test inserting a line at specific position."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool.insert(rel_path, 2, "inserted line")
        assert result.error is None
        lines = sample_file.read_text().splitlines()
        assert lines[1] == "inserted line"

    @pytest.mark.asyncio
    async def test_delete_lines(self, file_tool, sample_file, temp_workspace):
        """Test deleting specific lines."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool.delete_lines(rel_path, [2])
        assert result.error is None
        content = sample_file.read_text()
        assert "line 1" in content
        assert "line 2" not in content
        assert "line 3" in content

    @pytest.mark.asyncio
    async def test_rmdir(self, file_tool, temp_workspace):
        """Test removing an empty directory."""
        empty_dir = temp_workspace / "empty"
        empty_dir.mkdir()
        result = await file_tool.rmdir("empty")
        assert result.error is None
        assert not empty_dir.exists()

    @pytest.mark.asyncio
    async def test_unsupported_command(self, file_tool):
        """Test that unsupported commands return an error."""
        result = await file_tool(command="unsupported_cmd", path="test.txt")
        assert result.error is not None


class TestBashTool:
    """Integration tests for BashTool class."""

    @pytest.mark.asyncio
    async def test_simple_command(self, bash_tool):
        """Test executing a simple command."""
        result = await bash_tool(command="echo hello")
        assert result.output is not None
        assert "hello" in result.output

    @pytest.mark.asyncio
    async def test_command_with_exit_code(self, bash_tool):
        """Test command that fails."""
        result = await bash_tool(command="exit 1")
        assert result is not None

    @pytest.mark.asyncio
    async def test_list_sessions(self, bash_tool):
        """Test listing sessions."""
        await bash_tool(command="echo test")
        result = await bash_tool(list_sessions=True)
        assert result.output is not None or result.system is not None

    @pytest.mark.asyncio
    async def test_session_restart(self, bash_tool):
        """Test restarting a session."""
        await bash_tool(command="echo first", session=1)
        result = await bash_tool(restart=True, session=1)
        assert "restarted" in result.system.lower()

    @pytest.mark.asyncio
    async def test_command_timeout(self, bash_tool):
        """Test that long commands timeout."""
        result = await bash_tool(command="sleep 100", timeout=1)
        assert result.system is not None
        assert "timed out" in result.system.lower()

    @pytest.mark.asyncio
    async def test_multiple_sessions(self, bash_tool):
        """Test using multiple sessions."""
        result1 = await bash_tool(command="echo session1", session=1)
        result2 = await bash_tool(command="echo session2", session=2)
        assert "session1" in result1.output
        assert "session2" in result2.output

    @pytest.mark.asyncio
    async def test_check_session_not_found(self, bash_tool):
        """Test checking a session that doesn't exist."""
        result = await bash_tool(check_session=999)
        assert result.error is not None
        assert "not found" in result.error.lower()

    @pytest.mark.asyncio
    async def test_check_idle_session(self, bash_tool):
        """Test checking an idle session."""
        await bash_tool(command="echo test", session=1)
        await asyncio.sleep(0.5)
        result = await bash_tool(check_session=1)
        assert result.system is not None
        assert "no command running" in result.system.lower()

    @pytest.mark.asyncio
    async def test_session_reuse(self, bash_tool):
        """Test that sessions can be reused."""
        result1 = await bash_tool(command="echo first", session=1)
        result2 = await bash_tool(command="echo second", session=1)
        assert "first" in result1.output
        assert "second" in result2.output

    @pytest.mark.asyncio
    async def test_no_command_provided(self, bash_tool):
        """Test that error is raised when no command is provided."""
        result = await bash_tool(session=1)
        assert result.error is not None or result.system is not None


class TestValidation:
    """Tests for input validation."""

    def test_command_length_validation_passes(self):
        """Test that normal commands pass validation."""
        validate_command_length("echo hello")

    def test_command_length_validation_fails(self):
        """Test that very long commands are rejected."""
        long_command = "x" * 200000
        with pytest.raises(ToolError) as exc_info:
            validate_command_length(long_command)
        assert "too long" in str(exc_info.value.message).lower()


class TestBashSession:
    """Tests for _BashSession class."""

    @pytest.mark.asyncio
    async def test_session_starts(self):
        """Test that session starts properly."""
        session = _BashSession(session_id=1)
        await session.start()
        assert session._started is True
        session.stop()

    @pytest.mark.asyncio
    async def test_session_stops(self):
        """Test that session stops properly."""
        session = _BashSession(session_id=1)
        await session.start()
        session.stop()

    @pytest.mark.asyncio
    async def test_session_run_command(self):
        """Test running command in session."""
        session = _BashSession(session_id=1)
        await session.start()
        result = await session.run("echo test")
        assert "test" in result.output
        session.stop()

    @pytest.mark.asyncio
    async def test_session_properties(self):
        """Test session properties."""
        session = _BashSession(session_id=42)
        assert session.session_id == 42
        assert not session.is_running_command
        assert session.last_command == ""

    @pytest.mark.asyncio
    async def test_session_error_output_filtering(self):
        """Test that common error messages are filtered."""
        session = _BashSession(session_id=1)
        filtered = session._filter_error_output("failed to connect to the bus\nreal error\n")
        assert "real error" in filtered
        assert "failed to connect to the bus" not in filtered


class TestToolResult:
    """Tests for ToolResult class."""

    def test_tool_result_bool(self):
        """Test ToolResult boolean evaluation."""
        assert bool(ToolResult(output="test"))
        assert bool(ToolResult(error="error"))
        assert not bool(ToolResult())

    def test_tool_result_add(self):
        """Test combining ToolResults."""
        result1 = ToolResult(output="first")
        result2 = ToolResult(output=" second")
        combined = result1 + result2
        assert combined.output == "first second"

    def test_tool_result_replace(self):
        """Test replacing ToolResult fields."""
        result = ToolResult(output="test", error="error")
        new_result = result.replace(error=None)
        assert new_result.output == "test"
        assert new_result.error is None


class TestEdgeCases:
    """Tests for edge cases and error conditions."""

    @pytest.mark.asyncio
    async def test_read_binary_file(self, file_tool, temp_workspace):
        """Test reading a binary file."""
        binary_file = temp_workspace / "binary.bin"
        binary_file.write_bytes(b"\x00\x01\x02\x03")
        result = await file_tool.read("binary.bin", mode="binary")
        assert result.error is None
        assert result.system == "binary"

    @pytest.mark.asyncio
    async def test_write_binary_file(self, file_tool, temp_workspace):
        """Test writing a binary file."""
        import base64
        content = base64.b64encode(b"\x00\x01\x02\x03").decode()
        result = await file_tool.write("binary.bin", content, mode="binary")
        assert result.error is None
        assert (temp_workspace / "binary.bin").read_bytes() == b"\x00\x01\x02\x03"

    @pytest.mark.asyncio
    async def test_grep_recursive(self, file_tool, sample_dir, temp_workspace):
        """Test recursive grep in directory."""
        rel_path = str(sample_dir.relative_to(temp_workspace))
        result = await file_tool.grep(pattern="content", path=rel_path, recursive=True)
        assert result.error is None
        assert "content 1" in result.output or "content 2" in result.output

    @pytest.mark.asyncio
    async def test_grep_without_recursive_on_dir(self, file_tool, sample_dir, temp_workspace):
        """Test that grep fails on directory without recursive flag."""
        rel_path = str(sample_dir.relative_to(temp_workspace))
        result = await file_tool(command="grep", pattern="content", path=rel_path, recursive=False)
        assert result.error is not None
        assert "recursive" in result.error.lower()

    @pytest.mark.asyncio
    async def test_replace_not_found(self, file_tool, sample_file, temp_workspace):
        """Test replacing a string that doesn't exist."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool(command="replace", path=rel_path, old_str="nonexistent", new_str="replacement")
        assert result.error is not None
        assert "not found" in result.error.lower()

    @pytest.mark.asyncio
    async def test_replace_multiple_without_flag(self, file_tool, temp_workspace):
        """Test replacing multiple occurrences without all_occurrences flag."""
        test_file = temp_workspace / "test.txt"
        test_file.write_text("foo foo")
        result = await file_tool(command="replace", path="test.txt", old_str="foo", new_str="bar", all_occurrences=False)
        assert result.error is not None
        assert "multiple occurrences" in result.error.lower()

    @pytest.mark.asyncio
    async def test_undo_without_history(self, file_tool, sample_file, temp_workspace):
        """Test undo when there's no history."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool(command="undo", path=rel_path)
        assert result.error is not None
        assert "no undo history" in result.error.lower()

    @pytest.mark.asyncio
    async def test_delete_directory_recursive(self, file_tool, sample_dir, temp_workspace):
        """Test deleting a directory recursively."""
        rel_path = str(sample_dir.relative_to(temp_workspace))
        result = await file_tool.delete(rel_path, recursive=True)
        assert result.error is None
        assert not sample_dir.exists()

    @pytest.mark.asyncio
    async def test_view_directory(self, file_tool, sample_dir, temp_workspace):
        """Test viewing a directory."""
        rel_path = str(sample_dir.relative_to(temp_workspace))
        result = await file_tool.view(rel_path)
        assert result.error is None
        assert "file1.txt" in result.output
        assert "file2.txt" in result.output

    @pytest.mark.asyncio
    async def test_insert_out_of_range(self, file_tool, sample_file, temp_workspace):
        """Test inserting at invalid line number."""
        rel_path = str(sample_file.relative_to(temp_workspace))
        result = await file_tool(command="insert", path=rel_path, line=100, text="text")
        assert result.error is not None
        assert "out of range" in result.error.lower()
