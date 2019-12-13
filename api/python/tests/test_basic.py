import pytest
from heedy import App


def test_basics():
    a = App("testkey")
    assert a.owner.username == "test"

    a.owner.name = "Myname"
    assert a.owner.name == "Myname"

    assert len(a.objects())==0

    o = a.objects.create("myobj",{"schema": {"type": "number"}})
    assert o.name =="myobj"
    assert o.type == "stream"
    assert len(a.objects()) == 1

    assert o== a.objects[o.id]

    assert o.length()==0
    o.append(2)
    assert o.length()==1
    d = o[:]
    assert len(d)==1
    assert d[0]["d"] == 2
    o.remove() # Clear the stream
    assert len(o)==0

    o.delete()
    assert len(a.objects())==0
    # assert len(a.owner.apps())==1


def test_notifications():
    a = App("testkey")

    assert len(a.notifications())==0
    a.notify("hi","hello")
    assert len(a.notifications())==1
    a.notifications.delete("hi")
    assert len(a.notifications())==0

@pytest.mark.asyncio
async def test_basics_async():
    a = App("testkey", session="async")
    
    await (await a.owner).update(name= "Myname2")
    assert await (await a.owner).name == "Myname2"

    assert len(await a.objects())==0

    o = await a.objects.create("myobj2",{"schema": {"type": "number"}})
    assert o.name =="myobj2"
    assert o.type == "stream"
    assert len(await a.objects()) == 1

    assert o== await a.objects[o.id]

    await o.delete()


    # assert len(await (await a.owner).apps())==1